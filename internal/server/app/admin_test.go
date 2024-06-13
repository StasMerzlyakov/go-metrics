package app_test

import (
	"context"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/app"
	"github.com/StasMerzlyakov/go-metrics/internal/server/app/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPinger := mocks.NewMockPinger(ctrl)

	mockPinger.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)

	adminApp := app.NewAdminApp(mockPinger)
	require.NotNil(t, adminApp)

	err := adminApp.Ping(context.Background())
	require.NoError(t, err)

}
