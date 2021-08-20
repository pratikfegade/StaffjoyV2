package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/golang/protobuf/ptypes/empty"
	pb "v2.staffjoy.com/company"
	"v2.staffjoy.com/helpers"
)

func (s *companyServer) ListWorkers(ctx context.Context, req *pb.WorkerListRequest) (*pb.Workers, error) {
	if s.use_caching {
		s.lw_rw_lock.RLock()
		res, ok := s.lw_cache[req.TeamUuid]
		s.lw_rw_lock.RUnlock()
		if ok {
			return res, nil
		}
	}
	// Prep
	_, _, err := getAuth(ctx)

	if _, err = s.GetTeam(ctx, &pb.GetTeamRequest{CompanyUuid: req.CompanyUuid, Uuid: req.TeamUuid}); err != nil {
		return nil, err
	}

	res := &pb.Workers{CompanyUuid: req.CompanyUuid, TeamUuid: req.TeamUuid}

	rows, err := s.db.Query("select user_uuid from worker where team_uuid=?", req.TeamUuid)
	if err != nil {
		return nil, s.internalError(err, "unable to query database")
	}

	for rows.Next() {
		var userUUID string
		if err := rows.Scan(&userUUID); err != nil {
			return nil, s.internalError(err, "Error scanning database")
		}
		e, err := s.GetDirectoryEntry(ctx, &pb.DirectoryEntryRequest{CompanyUuid: req.CompanyUuid, UserUuid: userUUID})
		if err != nil {
			return nil, err
		}
		res.Workers = append(res.Workers, *e)
	}

	if s.use_caching {
		s.lw_rw_lock.Lock()
		s.lw_cache[req.TeamUuid] = res
		s.lw_rw_lock.Unlock()
	}
	return res, nil
}

func (s *companyServer) ListWorkers_UpdateAccount_Handler(ctx context.Context, req *pb.WorkerOfRequest) ( *pb.WorkerOfRequest, error) {
	s.logger.Info("[ListWorkers_UpdateAccount_Handler] Called")
	rows, err := s.db.Query("select team_uuid from worker where user_uuid=?", req.UserUuid)
	if err != nil {
		return req, s.internalError(err, "Unable to query database for invalidating LW cache")
	}

	for rows.Next() {
		var teamUUID string
		if err := rows.Scan(&teamUUID); err != nil {
			return req, s.internalError(err, "err scanning database for invalidating LW cache")
		}
		s.lw_rw_lock.Lock()
		delete(s.lw_cache, teamUUID)
		s.lw_rw_lock.Unlock()
		s.logger.Info("[ListWorkers_UpdateAccount_Handler] Invalidated cache")
		s.logger.Info(teamUUID)
	}
	return req, nil
}

func (s *companyServer) ListWorkers_DeleteWorker_Handler(team_uuid string) ( error) {

	s.lw_rw_lock.Lock()
	delete(s.lw_cache, team_uuid)
	s.lw_rw_lock.Unlock()
	s.logger.Info("[ListWorkers_DeleteWorker_Handler] Invalidated cache")
	s.logger.Info(team_uuid)
	return nil
}

func (s *companyServer) GetWorker(ctx context.Context, req *pb.Worker) (*pb.DirectoryEntry, error) {
	_, _, err := getAuth(ctx)
	if _, err = s.GetTeam(ctx, &pb.GetTeamRequest{CompanyUuid: req.CompanyUuid, Uuid: req.TeamUuid}); err != nil {
		return nil, err
	}

	var exists bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM worker WHERE (team_uuid=? AND user_uuid=?))", req.TeamUuid, req.UserUuid).Scan(&exists)
	if err != nil {
		return nil, s.internalError(err, "failed to query database")
	} else if !exists {
		return nil, grpc.Errorf(codes.NotFound, "worker relationship not found")
	}
	return s.GetDirectoryEntry(ctx, &pb.DirectoryEntryRequest{CompanyUuid: req.CompanyUuid, UserUuid: req.UserUuid})
}

func (s *companyServer) DeleteWorker(ctx context.Context, req *pb.Worker) (*empty.Empty, error) {
	md, _, err := getAuth(ctx)

	if _, err = s.GetWorker(ctx, req); err != nil {
		return nil, err
	}
	if _, err = s.db.Exec("DELETE from worker where (team_uuid=? AND user_uuid=?) LIMIT 1", req.TeamUuid, req.UserUuid); err != nil {
		return nil, s.internalError(err, "failed to query database")
	}

	if s.use_caching {
		err = s.ListWorkers_DeleteWorker_Handler(req.TeamUuid)
		err = s.GetWorkerOf_DeleteWorker_Handler(req.UserUuid)
	}

	al := newAuditEntry(md, "worker", req.UserUuid, req.CompanyUuid, req.TeamUuid)
	al.Log(logger, "removed worker")
	go helpers.TrackEventFromMetadata(md, "worker_deleted", ServiceName)
	return &empty.Empty{}, nil
}

func (s *companyServer) GetWorkerOf(ctx context.Context, req *pb.WorkerOfRequest) (*pb.WorkerOfList, error) {
	if s.use_caching {
		s.gwe_rw_lock.RLock()
		res, ok := s.gwe_cache[req.UserUuid]
		s.gwe_rw_lock.RUnlock()
		if ok {
			return res, nil
		}
	}
	_, _, err := getAuth(ctx)

	res := &pb.WorkerOfList{UserUuid: req.UserUuid}

	rows, err := s.db.Query("select worker.team_uuid, team.company_uuid from worker JOIN team ON team.uuid=worker.team_uuid where worker.user_uuid=?", req.UserUuid)
	if err != nil {
		return nil, s.internalError(err, "Unable to query database")
	}

	for rows.Next() {
		var teamUUID, companyUUID string
		if err := rows.Scan(&teamUUID, &companyUUID); err != nil {
			return nil, s.internalError(err, "err scanning database")
		}
		t, err := s.GetTeam(ctx, &pb.GetTeamRequest{Uuid: teamUUID, CompanyUuid: companyUUID})
		if err != nil {
			return nil, err
		}
		res.Teams = append(res.Teams, *t)
	}

	if s.use_caching {
		s.gwe_rw_lock.Lock()
		s.gwe_cache[req.UserUuid] = res
		s.gwe_rw_lock.Unlock()
	}
	return res, nil

}

func (s *companyServer) GetWorkerOf_UpdateTeam_Handler(uuid string) (error) {
	rows, err := s.db.Query("select worker.user_uuid from worker where worker.team_uuid=?", uuid)
	if err != nil {
		return s.internalError(err, "Unable to query database for invalidating GWE cache")
	}

	for rows.Next() {
		var userUUID string
		if err := rows.Scan(&userUUID); err != nil {
			return s.internalError(err, "err scanning database for invalidating GWE cache")
		}
		s.gwe_rw_lock.Lock()
		delete(s.gwe_cache, userUUID)
		s.gwe_rw_lock.Unlock()
		s.logger.Info("[GetWorkerOf_UpdateTeam_Handler] Invalidated cache")
		s.logger.Info(userUUID)
	}
	return nil
}

func (s *companyServer) GetWorkerOf_DeleteWorker_Handler(uuid string) (error) {
	s.gwe_rw_lock.Lock()
	delete(s.gwe_cache, uuid)
	s.gwe_rw_lock.Unlock()
	s.logger.Info("[GetWorkerOf_UpdateTeam_Handler] Invalidated cache")
	s.logger.Info(uuid)
	return nil
}

func (s *companyServer) CreateWorker(ctx context.Context, req *pb.Worker) (*pb.DirectoryEntry, error) {
	md, _, err := getAuth(ctx)
	if _, err := s.GetTeam(ctx, &pb.GetTeamRequest{CompanyUuid: req.CompanyUuid, Uuid: req.TeamUuid}); err != nil {
		return nil, err

	}
	e, err := s.GetDirectoryEntry(ctx, &pb.DirectoryEntryRequest{CompanyUuid: req.CompanyUuid, UserUuid: req.UserUuid})
	if err != nil {
		return nil, err
	}

	_, err = s.GetWorker(ctx, req)
	if err == nil {
		return nil, grpc.Errorf(codes.AlreadyExists, "user is already a worker")
	} else if grpc.Code(err) != codes.NotFound {
		return nil, s.internalError(err, "an unknown error occurred while checking for existing worker relationships")
	}

	_, err = s.db.Exec("INSERT INTO worker (team_uuid, user_uuid) values (?, ?)", req.TeamUuid, req.UserUuid)
	if err != nil {
		return nil, s.internalError(err, "failed to query database")
	}
	al := newAuditEntry(md, "worker", req.UserUuid, req.CompanyUuid, req.TeamUuid)
	al.Log(logger, "added worker")
	go helpers.TrackEventFromMetadata(md, "worker_created", ServiceName)

	return e, nil
}
