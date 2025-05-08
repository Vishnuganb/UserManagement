package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"UserManagement/internal/util"
)

// The individual tester for a single test function
func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		FirstName: util.RandomName(),
		LastName:  util.RandomName(),
		Email:     util.RandomEmail(),
		Phone:     sql.NullString{String: util.RandomPhone(), Valid: true}, // Valid: true means "this is NOT NULL".
		Age:       sql.NullInt32{Int32: util.RandomAge(), Valid: true},
		Status:    sql.NullString{String: util.RandomStatus(), Valid: true},
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, arg.FirstName, user.FirstName)
	require.Equal(t, arg.LastName, user.LastName)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Phone, user.Phone)
	require.Equal(t, arg.Age, user.Age)
	require.Equal(t, arg.Status, user.Status)

	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.UpdatedAt)
	require.NotZero(t, user.UserID)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.UserID)
	require.NoError(t, err)
	require.Equal(t, user1.UserID, user2.UserID)
	require.Equal(t, user1.FirstName, user2.FirstName)
	require.Equal(t, user1.LastName, user2.LastName)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Phone, user2.Phone)
	require.Equal(t, user1.Age, user2.Age)
	require.Equal(t, user1.Status, user2.Status)
	require.NotZero(t, user1.CreatedAt)
	require.NotZero(t, user1.UpdatedAt)
}

func TestUpdateUser(t *testing.T) {
	user1 := createRandomUser(t)
	arg := UpdateUserParams{
		UserID:    user1.UserID,
		FirstName: sql.NullString{String: util.RandomName(), Valid: true},
		LastName:  sql.NullString{String: util.RandomName(), Valid: true},
		Email:     sql.NullString{String: util.RandomEmail(), Valid: true},
		Phone:     sql.NullString{String: util.RandomPhone(), Valid: true}, // Valid: true means "this is NOT NULL".
		Age:       sql.NullInt32{Int32: util.RandomAge(), Valid: true},
		Status:    sql.NullString{String: util.RandomStatus(), Valid: true},
	}
	user2, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, arg.UserID, user2.UserID)
	require.Equal(t, arg.FirstName.String, user2.FirstName)
	require.Equal(t, arg.LastName.String, user2.LastName)
	require.Equal(t, arg.Email.String, user2.Email)
	require.Equal(t, arg.Phone.String, user2.Phone.String)
	require.Equal(t, arg.Age.Int32, user2.Age.Int32)
	require.Equal(t, arg.Status.String, user2.Status.String)
	require.NotZero(t, user2.CreatedAt)
	require.NotZero(t, user2.UpdatedAt)
}

func TestDeleteUser(t *testing.T) {
	user1 := createRandomUser(t)
	err := testQueries.DeleteUser(context.Background(), user1.UserID)
	require.NoError(t, err)
	user2, err := testQueries.GetUser(context.Background(), user1.UserID)
	require.Error(t, err)
	require.Equal(t, err, sql.ErrNoRows)
	require.Empty(t, user2)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomUser(t)
	}

	users, err := testQueries.ListUsers(context.Background())
	require.NoError(t, err)
	for _, user := range users {
		require.NotEmpty(t, user)
	}
}
