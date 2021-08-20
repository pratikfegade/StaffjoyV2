package main

import (
	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "v2.staffjoy.com/company"
	"v2.staffjoy.com/crypto"
	"v2.staffjoy.com/helpers"
)

func (s *companyServer) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.Job, error) {
	md, _, err := getAuth(ctx)

	if _, err = s.GetTeam(ctx, &pb.GetTeamRequest{Uuid: req.TeamUuid, CompanyUuid: req.CompanyUuid}); err != nil {
		return nil, err
	}

	if err = validColor(req.Color); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "Invalid color")
	}

	uuid, err := crypto.NewUUID()
	if err != nil {
		return nil, s.internalError(err, "Cannot generate a uuid")
	}
	j := &pb.Job{Uuid: uuid.String(), Name: req.Name, Color: req.Color, CompanyUuid: req.CompanyUuid, TeamUuid: req.TeamUuid}

	if err = s.dbMap.Insert(j); err != nil {
		return nil, s.internalError(err, "could not create job")
	}

	al := newAuditEntry(md, "job", j.Uuid, j.CompanyUuid, j.TeamUuid)
	al.UpdatedContents = j
	al.Log(logger, "created job")
	go helpers.TrackEventFromMetadata(md, "job_created", ServiceName)

	return j, nil
}

func (s *companyServer) ListJobs(ctx context.Context, req *pb.JobListRequest) (*pb.JobList, error) {
	if s.use_caching {
		s.lj_rw_lock.RLock()
		res, ok := s.lj_cache[req.TeamUuid]
		s.lj_rw_lock.RUnlock()

		if ok {
			return res, nil
		}
	}
	_, _, err := getAuth(ctx)
	if _, err = s.GetTeam(ctx, &pb.GetTeamRequest{Uuid: req.TeamUuid, CompanyUuid: req.CompanyUuid}); err != nil {
		return nil, err
	}

	res := &pb.JobList{}
	rows, err := s.db.Query("select uuid from job where team_uuid=?", req.TeamUuid)
	if err != nil {
		return nil, s.internalError(err, "unable to query database")
	}

	for rows.Next() {
		r := &pb.GetJobRequest{CompanyUuid: req.CompanyUuid, TeamUuid: req.TeamUuid}
		if err := rows.Scan(&r.Uuid); err != nil {
			return nil, s.internalError(err, "error scanning database")
		}

		var j *pb.Job
		if j, err = s.GetJob(ctx, r); err != nil {
			return nil, err
		}
		res.Jobs = append(res.Jobs, *j)
	}

	if s.use_caching {
		s.lj_rw_lock.Lock()
		s.lj_cache[req.TeamUuid] = res
		s.lj_rw_lock.Unlock()
	}
	return res, nil
}

func (s *companyServer) ListJobs_UpdateJob_Handler(uuid string) (error) {
	rows, err := s.db.Query("select team_uuid from job where uuid=?", uuid)
	if err != nil {
		return s.internalError(err, "Unable to query database for invalidating LJ cache")
	}

	for rows.Next() {
		var teamUUID string
		if err := rows.Scan(&teamUUID); err != nil {
			return s.internalError(err, "err scanning database for invalidating LJ cache")
		}
		s.lj_rw_lock.Lock()
		delete(s.lj_cache, teamUUID)
		s.lj_rw_lock.Unlock()
		s.logger.Info("[ListJobs_UpdateJob_Handler] Invalidated LJ cache")
		s.logger.Info(teamUUID)
	}
	return nil
}

func (s *companyServer) ListJobs_UpdateJobTeamUUID_Handler(uuid string, old_team string, new_team string) (error) {
	s.lj_rw_lock.Lock()
	delete(s.lj_cache, old_team)
	delete(s.lj_cache, new_team)
	s.lj_rw_lock.Unlock()
	s.logger.Info("[ListJobs_UpdateJobTeamUUID_Handler] Invalidated LJ cache")
	s.logger.Info(old_team)
	s.logger.Info(new_team)
	return nil
}

func (s *companyServer) GetJob(ctx context.Context, req *pb.GetJobRequest) (*pb.Job, error) {
	_, _, err := getAuth(ctx)

	if _, err = s.GetTeam(ctx, &pb.GetTeamRequest{Uuid: req.TeamUuid, CompanyUuid: req.CompanyUuid}); err != nil {
		return nil, err
	}

	obj, err := s.dbMap.Get(pb.Job{}, req.Uuid)
	if err != nil {
		return nil, s.internalError(err, "unable to query database")
	} else if obj == nil {
		return nil, grpc.Errorf(codes.NotFound, "job not found")
	}
	j := obj.(*pb.Job)
	j.CompanyUuid = req.CompanyUuid
	j.TeamUuid = req.TeamUuid
	return j, nil
}

func (s *companyServer) UpdateJob(ctx context.Context, req *pb.Job) (*pb.Job, error) {
	_, _, err := getAuth(ctx)

	if _, err = s.GetTeam(ctx, &pb.GetTeamRequest{Uuid: req.TeamUuid, CompanyUuid: req.CompanyUuid}); err != nil {
		return nil, grpc.Errorf(codes.NotFound, "Company and team path not found")
	}

	if err = validColor(req.Color); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "Invalid color")
	}

	orig, err := s.GetJob(ctx, &pb.GetJobRequest{CompanyUuid: req.CompanyUuid, TeamUuid: req.TeamUuid, Uuid: req.Uuid})
	if err != nil {
		return nil, err
	}

	if _, err := s.db.Exec("update job set team_uuid=?, name=?, archived=?, color=?",
		req.TeamUuid, req.Name, req.Archived, req.Color); err != nil {
		return nil, s.internalError(err, "could not update the job")
	}

	if s.use_caching {
		if orig.TeamUuid != req.TeamUuid {
			s.ListJobs_UpdateJobTeamUUID_Handler(req.Uuid, orig.TeamUuid, req.TeamUuid)
		} else {
			s.ListJobs_UpdateJob_Handler(req.Uuid)
		}
	}

	return req, nil
}
