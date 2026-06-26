package overlays

func init() {
	// An AWS RDS database instance is slow to provision (minutes) and its endpoint
	// (Endpoint.Address / Endpoint.Port) is only assigned once the instance is
	// available, while its identifier/ARN are known early.
	RegisterStabiliseRequired("aws/rds/dbInstance")
}
