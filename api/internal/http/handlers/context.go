package handlers

import "context"

type contextKey string

const (
	contextUserIDKey    contextKey = "user_id"
	contextCompanyIDKey contextKey = "company_id"
	contextRoleKey      contextKey = "role"
)

func WithAuthContext(ctx context.Context, userID, companyID, role string) context.Context {
	ctx = context.WithValue(ctx, contextUserIDKey, userID)
	ctx = context.WithValue(ctx, contextCompanyIDKey, companyID)
	ctx = context.WithValue(ctx, contextRoleKey, role)

	return ctx
}

func UserIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(contextUserIDKey).(string)
	return value
}

func CompanyIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(contextCompanyIDKey).(string)
	return value
}

func RoleFromContext(ctx context.Context) string {
	value, _ := ctx.Value(contextRoleKey).(string)
	return value
}
