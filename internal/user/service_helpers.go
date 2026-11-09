package user

func makeProfile(foundUser User) Profile {
	return Profile{
		ID:        foundUser.ID,
		FullName:  foundUser.FullName,
		Email:     foundUser.Email,
		Role:      foundUser.Role,
		AvatarKey: foundUser.AvatarKey,
		CreatedAt: foundUser.CreatedAt,
		UpdatedAt: foundUser.UpdatedAt,
	}
}
