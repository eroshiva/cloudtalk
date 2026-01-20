package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/lock,sql/versioned-migration,sql/execquery,intercept ./schema
