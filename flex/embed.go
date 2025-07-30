package flex

import "embed"

//go:embed examples/*
var examples embed.FS

//go:embed descriptions/*
var descriptions embed.FS
