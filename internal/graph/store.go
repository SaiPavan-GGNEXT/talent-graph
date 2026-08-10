package graph

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
)

// ErrUnavailable is returned when the database cannot be reached, so the
// HTTP layer can translate it into a 503 instead of a generic 500.
var ErrUnavailable = errors.New("graph database unavailable")

// Store wraps the Neo4j driver and exposes the application's queries.
// All Cypher lives in this package and every query is parameterised.
type Store struct {
	driver neo4j.DriverWithContext
}

// New connects to the database and verifies connectivity once at startup.
func New(ctx context.Context, uri, user, password string) (*Store, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""),
		func(c *config.Config) {
			// Fail fast when the database is down so the API can answer 503
			// promptly instead of retrying for the default 30 seconds.
			c.MaxTransactionRetryTime = 5 * time.Second
			// The CognoDB free tier allows up to 200 connections; stay well
			// under it while still supporting healthy concurrency.
			c.MaxConnectionPoolSize = 50
			c.ConnectionAcquisitionTimeout = 10 * time.Second
		})
	if err != nil {
		return nil, err
	}
	verifyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := driver.VerifyConnectivity(verifyCtx); err != nil {
		driver.Close(ctx)
		return nil, errors.Join(ErrUnavailable, err)
	}
	return &Store{driver: driver}, nil
}

func (s *Store) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

// Ping reports whether the database is currently reachable.
func (s *Store) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.driver.VerifyConnectivity(ctx); err != nil {
		return errors.Join(ErrUnavailable, err)
	}
	return nil
}

// read runs a read transaction and collects the produced records.
func (s *Store) read(ctx context.Context, cypher string, params map[string]any) ([]*neo4j.Record, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		return res.Collect(ctx)
	})
	if err != nil {
		return nil, wrapConnErr(err)
	}
	return result.([]*neo4j.Record), nil
}

// write runs a write transaction; used by the seed loader.
func (s *Store) write(ctx context.Context, cypher string, params map[string]any) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		return res.Consume(ctx)
	})
	return wrapConnErr(err)
}

// Exec runs a single write statement; used by the seed loader.
func (s *Store) Exec(ctx context.Context, cypher string, params map[string]any) error {
	return s.write(ctx, cypher, params)
}

// wrapConnErr tags connectivity-class failures with ErrUnavailable.
func wrapConnErr(err error) error {
	if err == nil {
		return nil
	}
	// TransactionExecutionLimit means the managed-transaction retries were
	// exhausted — in practice that is the driver failing to reach the server.
	if neo4j.IsConnectivityError(err) || neo4j.IsTransactionExecutionLimit(err) {
		return errors.Join(ErrUnavailable, err)
	}
	return err
}
