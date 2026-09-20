package service

func hasMembershipRole(roles []string) bool {
	for _, role := range roles {
		switch role {
		case "member", "officer", "probationary", "president", "vice_president", "minister", "vice_minister", "center_director", "vice_center_director", "honorary":
			return true
		}
	}
	return false
}

func applicationGates(roles []string, ident *MemberIdentity) (canApplyMember, canApplyOfficer bool) {
	profileCurrent := ident != nil && (ident.Status == 0 || ident.Status == 3) && ident.MemberType >= 1
	canApplyMember = !hasMembershipRole(roles) && !profileCurrent
	canApplyOfficer = ident != nil && ident.Status == 0 && ident.MemberType >= 1 && ident.MemberType < 2
	return
}
