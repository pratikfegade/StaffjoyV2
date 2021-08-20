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

func (s *companyServer) CreateTeam(ctx context.Context, req *pb.CreateTeamRequest) (*pb.Team, error) {
	md, _, err := getAuth(ctx)

	c, err := s.GetCompany(ctx, &pb.GetCompanyRequest{Uuid: req.CompanyUuid})
	if err != nil {
		return nil, grpc.Errorf(codes.NotFound, "Company with specified id not found")
	}

	if req.DayWeekStarts == "" {
		req.DayWeekStarts = c.DefaultDayWeekStarts
	} else if req.DayWeekStarts, err = sanitizeDayOfWeek(req.DayWeekStarts); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid DefaultDayWeekStarts")
	}
	if req.Timezone == "" {
		req.Timezone = c.DefaultTimezone
	} else if err = validTimezone(req.Timezone); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid timezone")
	}

	if err = validColor(req.Color); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid color")
	}

	uuid, err := crypto.NewUUID()
	if err != nil {
		return nil, s.internalError(err, "cannot generate a uuid")
	}
	t := &pb.Team{Uuid: uuid.String(), CompanyUuid: req.CompanyUuid, Name: req.Name, DayWeekStarts: req.DayWeekStarts, Timezone: req.Timezone, Color: req.Color}

	if err = s.dbMap.Insert(t); err != nil {
		return nil, s.internalError(err, "could not create team")
	}

	al := newAuditEntry(md, "team", t.Uuid, req.CompanyUuid, t.Uuid)
	al.UpdatedContents = t
	al.Log(logger, "created team")
	go helpers.TrackEventFromMetadata(md, "team_created", ServiceName)

	return t, nil
}

func (s *companyServer) ListTeams(ctx context.Context, req *pb.TeamListRequest) (*pb.TeamList, error) {
	_, _, err := getAuth(ctx)

	if _, err = s.GetCompany(ctx, &pb.GetCompanyRequest{Uuid: req.CompanyUuid}); err != nil {
		return nil, err
	}

	res := &pb.TeamList{}
	rows, err := s.db.Query("select uuid from team where company_uuid=?", req.CompanyUuid)
	if err != nil {
		return nil, s.internalError(err, "unable to query database")
	}

	for rows.Next() {
		r := &pb.GetTeamRequest{CompanyUuid: req.CompanyUuid}
		if err := rows.Scan(&r.Uuid); err != nil {
			return nil, s.internalError(err, "error scanning database")
		}

		var t *pb.Team
		if t, err = s.GetTeam(ctx, r); err != nil {
			return nil, err
		}
		res.Teams = append(res.Teams, *t)
	}
	return res, nil
}

func (s *companyServer) GetTeam(ctx context.Context, req *pb.GetTeamRequest) (*pb.Team, error) {
	_, _, err := getAuth(ctx)


	if _, err = s.GetCompany(ctx, &pb.GetCompanyRequest{Uuid: req.CompanyUuid}); err != nil {
		return nil, err
	}

	obj, err := s.dbMap.Get(pb.Team{}, req.Uuid)
	if err != nil {
		return nil, s.internalError(err, "unable to query database")
	} else if obj == nil {
		return nil, grpc.Errorf(codes.NotFound, "team not found")
	}
	t := obj.(*pb.Team)
	t.CompanyUuid = req.CompanyUuid
	return t, nil

}

func (s *companyServer) UpdateTeam(ctx context.Context, req *pb.Team) (*pb.Team, error) {
	_, _, err := getAuth(ctx)

	if _, err = s.GetCompany(ctx, &pb.GetCompanyRequest{Uuid: req.CompanyUuid}); err != nil {
		return nil, err
	}

	if req.DayWeekStarts, err = sanitizeDayOfWeek(req.DayWeekStarts); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "Invalid DefaultDayWeekStarts")
	}
	if err = validTimezone(req.Timezone); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "Invalid timezone")
	}
	if err = validColor(req.Color); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "Invalid color")
	}

	_, err = s.GetTeam(ctx, &pb.GetTeamRequest{CompanyUuid: req.CompanyUuid, Uuid: req.Uuid})
	if err != nil {
		return nil, err
	}

	if _, err := s.db.Exec("update team set company_uuid=?, name=?, archived=?, timezone=?, day_week_starts=?, color=?",
		req.CompanyUuid, req.Name, req.Archived, req.Timezone, req.DayWeekStarts, req.Color); err != nil {
		return nil, s.internalError(err, "could not update the team ")
	}

	if s.use_caching {
		err = s.GetWorkerOf_UpdateTeam_Handler(req.Uuid)
	}

	return req, err
}

func (s *companyServer) GetWorkerTeamInfo(ctx context.Context, req *pb.Worker) (*pb.Worker, error) {
	_, _, err := getAuth(ctx)

	teamUUID := ""
	q := "select team_uuid from worker where user_uuid = ?;"
	err = s.db.QueryRow(q, req.UserUuid).Scan(&teamUUID)
	if err != nil {
		logger.Debugf("get team -- %v", err)
		return nil, grpc.Errorf(codes.InvalidArgument, "Invalid user")
	}

	companyUUID := ""
	q = "select company_uuid from team where uuid = ?;"
	err = s.db.QueryRow(q, teamUUID).Scan(&companyUUID)
	if err != nil {
		logger.Debugf("get team -- %v", err)
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid company")
	}

	w := &pb.Worker{
		CompanyUuid: companyUUID,
		TeamUuid:    teamUUID,
		UserUuid:    req.UserUuid,
	}

	return w, nil
}
