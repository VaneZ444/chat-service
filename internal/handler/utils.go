package handler

import (
	"context"
	"log/slog"
	"strconv"

	"google.golang.org/grpc/metadata"
)

func GetUserIDFromCtx(ctx context.Context) int64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0
	}

	ids := md.Get("user_id")
	if len(ids) == 0 {
		return 0
	}

	userID, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil {
		return 0
	}

	return userID
}

func GetUserNicknameFromCtx(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	nicks := md.Get("nickname")
	if len(nicks) == 0 {
		return ""
	}

	slog.Info("ctx nickname", "nickname", nicks[0])
	return nicks[0]
}
