package main

import (
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	pb "v2.staffjoy.com/company"
	"v2.staffjoy.com/helpers"
)

func (s *companyServer) ListAdmins(ctx context.Context, req *pb.AdminListRequest) (*pb.Admins, error) {
	_, _, err := getAuth(ctx)

	if _, err = s.GetCompany(ctx, &pb.GetCompanyRequest{Uuid: req.CompanyUuid}); err != nil {
		return nil, err
	}

	res := &pb.Admins{CompanyUuid: req.CompanyUuid}

	rows, err := s.db.Query("select user_uuid from admin where company_uuid=?", req.CompanyUuid)
	if err != nil {
		return nil, s.internalError(err, "Unable to query database")
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
		res.Admins = append(res.Admins, *e)
	}
	return res, nil
}

func (s *companyServer) GetAdmin(ctx context.Context, req *pb.DirectoryEntryRequest) (*pb.DirectoryEntry, error) {
	_, _, err := getAuth(ctx)

	if _, err = s.GetCompany(ctx, &pb.GetCompanyRequest{Uuid: req.CompanyUuid}); err != nil {
		return nil, err
	}

	var exists bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM admin WHERE (company_uuid=? AND user_uuid=?))",
		req.CompanyUuid, req.UserUuid).Scan(&exists)
	if err != nil {
		return nil, s.internalError(err, "failed to query database")
	} else if !exists {
		return nil, grpc.Errorf(codes.NotFound, "admin relationship not found")
	}
	return s.GetDirectoryEntry(ctx, req)
}

func (s *companyServer) DeleteAdmin(ctx context.Context, req *pb.DirectoryEntryRequest) (*empty.Empty, error) {
	md, _, err := getAuth(ctx)
	_, err = s.GetAdmin(ctx, req)
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec("DELETE from admin where (company_uuid=? AND user_uuid=?) LIMIT 1", req.CompanyUuid, req.UserUuid)
	if err != nil {
		return nil, s.internalError(err, "failed to query database")
	}
	al := newAuditEntry(md, "admin", req.UserUuid, req.CompanyUuid, "")
	al.Log(logger, "removed admin")
	go helpers.TrackEventFromMetadata(md, "admin_deleted", ServiceName)
	return &empty.Empty{}, nil
}

func (s *companyServer) CreateAdmin(ctx context.Context, req *pb.DirectoryEntryRequest) (*pb.DirectoryEntry, error) {
	md, _, err := getAuth(ctx)
	_, err = s.GetAdmin(ctx, req)
	if err == nil {
		return nil, grpc.Errorf(codes.AlreadyExists, "user is already an admin")
	} else if grpc.Code(err) != codes.NotFound {
		return nil, s.internalError(err, "an unknown error occurred while checking existing relationships")
	}

	e, err := s.GetDirectoryEntry(ctx, req)
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec("INSERT INTO admin (company_uuid, user_uuid) values (?, ?)", req.CompanyUuid, req.UserUuid)
	if err != nil {
		return nil, s.internalError(err, "failed to query database")
	}
	al := newAuditEntry(md, "admin", req.UserUuid, req.CompanyUuid, "")
	al.Log(logger, "added admin")
	go helpers.TrackEventFromMetadata(md, "admin_created", ServiceName)

	return e, nil
}

func (s *companyServer) GetAdminOf(ctx context.Context, req *pb.AdminOfRequest) (*pb.AdminOfList, error) {
	_, _, err := getAuth(ctx)

	res := &pb.AdminOfList{UserUuid: req.UserUuid}

	rows, err := s.db.Query("select company_uuid from admin where user_uuid=?", req.UserUuid)
	if err != nil {
		return nil, s.internalError(err, "Unable to query database")
	}

	for rows.Next() {
		var companyUUID string
		if err := rows.Scan(&companyUUID); err != nil {
			return nil, s.internalError(err, "err scanning database")
		}
		c, err := s.GetCompany(ctx, &pb.GetCompanyRequest{Uuid: companyUUID})
		if err != nil {
			return nil, err
		}
		res.Companies = append(res.Companies, *c)
	}

	return res, nil
}
