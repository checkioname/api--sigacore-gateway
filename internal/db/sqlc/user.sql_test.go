package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/checkioname/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hash, _ := util.HashPassword(util.RandomOwner())
	name := util.RandomOwner()
	arg := CreateUserParams{
		Username:       name,
		HashedPassword: hash,
		FullName:       fmt.Sprintf(name, " ", util.RandomMoney()),
		Email:          fmt.Sprintf("%s@example.com", name),
	}

	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.FullName, user.FullName)

	require.NotZero(t, user.CreatedAt)

	return user
}
