package seed

import (
	"context"
	"fmt"
	"log"

	"talentgraph/internal/graph"
)

// Load wipes the database and loads the generated dataset using batched,
// parameterised UNWIND statements — one round-trip per entity type.
func Load(ctx context.Context, store *graph.Store) error {
	ds := Build()

	log.Println("wiping existing data…")
	if err := store.Exec(ctx, `MATCH (n) DETACH DELETE n`, nil); err != nil {
		return fmt.Errorf("wipe: %w", err)
	}

	// Uniqueness constraints (best effort — syntax support can vary by engine).
	for _, c := range []string{
		`CREATE CONSTRAINT person_id IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE`,
		`CREATE CONSTRAINT skill_name IF NOT EXISTS FOR (s:Skill) REQUIRE s.name IS UNIQUE`,
		`CREATE CONSTRAINT project_id IF NOT EXISTS FOR (p:Project) REQUIRE p.id IS UNIQUE`,
		`CREATE CONSTRAINT client_name IF NOT EXISTS FOR (c:Client) REQUIRE c.name IS UNIQUE`,
	} {
		if err := store.Exec(ctx, c, nil); err != nil {
			log.Printf("constraint skipped (%v)", err)
		}
	}

	log.Printf("loading %d skills…", len(ds.Skills))
	skillRows := make([]map[string]any, len(ds.Skills))
	for i, s := range ds.Skills {
		skillRows[i] = map[string]any{"name": s.Name, "category": s.Category}
	}
	if err := store.Exec(ctx, `
		UNWIND $rows AS row
		MERGE (s:Skill {name: row.name})
		SET s.category = row.category`,
		map[string]any{"rows": skillRows}); err != nil {
		return fmt.Errorf("skills: %w", err)
	}

	log.Printf("loading %d clients…", len(ds.Clients))
	clientRows := make([]map[string]any, len(ds.Clients))
	for i, c := range ds.Clients {
		clientRows[i] = map[string]any{"name": c.Name, "industry": c.Industry}
	}
	if err := store.Exec(ctx, `
		UNWIND $rows AS row
		MERGE (c:Client {name: row.name})
		SET c.industry = row.industry`,
		map[string]any{"rows": clientRows}); err != nil {
		return fmt.Errorf("clients: %w", err)
	}

	log.Printf("loading %d people…", len(ds.People))
	personRows := make([]map[string]any, len(ds.People))
	for i, p := range ds.People {
		personRows[i] = map[string]any{
			"id": p.ID, "name": p.Name, "title": p.Title,
			"seniority": p.Seniority, "location": p.Location,
		}
	}
	if err := store.Exec(ctx, `
		UNWIND $rows AS row
		MERGE (p:Person {id: row.id})
		SET p.name = row.name, p.title = row.title,
		    p.seniority = row.seniority, p.location = row.location`,
		map[string]any{"rows": personRows}); err != nil {
		return fmt.Errorf("people: %w", err)
	}

	log.Printf("loading %d projects…", len(ds.Projects))
	projectRows := make([]map[string]any, len(ds.Projects))
	for i, pr := range ds.Projects {
		projectRows[i] = map[string]any{
			"id": pr.ID, "name": pr.Name, "client": pr.Client,
			"year": pr.Year, "status": pr.Status, "requires": pr.Requires,
		}
	}
	if err := store.Exec(ctx, `
		UNWIND $rows AS row
		MERGE (pr:Project {id: row.id})
		SET pr.name = row.name, pr.year = row.year, pr.status = row.status
		WITH pr, row
		MATCH (c:Client {name: row.client})
		MERGE (pr)-[:FOR_CLIENT]->(c)
		WITH pr, row
		UNWIND row.requires AS reqSkill
		MATCH (s:Skill {name: reqSkill})
		MERGE (pr)-[:REQUIRES]->(s)`,
		map[string]any{"rows": projectRows}); err != nil {
		return fmt.Errorf("projects: %w", err)
	}

	log.Printf("loading %d HAS_SKILL relationships…", len(ds.HasSkill))
	hsRows := make([]map[string]any, len(ds.HasSkill))
	for i, hs := range ds.HasSkill {
		hsRows[i] = map[string]any{
			"person": hs.PersonID, "skill": hs.Skill,
			"level": hs.Level, "years": hs.Years,
		}
	}
	if err := store.Exec(ctx, `
		UNWIND $rows AS row
		MATCH (p:Person {id: row.person}), (s:Skill {name: row.skill})
		MERGE (p)-[r:HAS_SKILL]->(s)
		SET r.level = row.level, r.years = row.years`,
		map[string]any{"rows": hsRows}); err != nil {
		return fmt.Errorf("has_skill: %w", err)
	}

	log.Printf("loading %d WORKED_ON relationships…", len(ds.WorkedOn))
	woRows := make([]map[string]any, len(ds.WorkedOn))
	for i, wo := range ds.WorkedOn {
		woRows[i] = map[string]any{
			"person": wo.PersonID, "project": wo.ProjectID, "role": wo.Role,
		}
	}
	if err := store.Exec(ctx, `
		UNWIND $rows AS row
		MATCH (p:Person {id: row.person}), (pr:Project {id: row.project})
		MERGE (p)-[r:WORKED_ON]->(pr)
		SET r.role = row.role`,
		map[string]any{"rows": woRows}); err != nil {
		return fmt.Errorf("worked_on: %w", err)
	}

	log.Printf("loading %d MENTORS relationships…", len(ds.Mentors))
	mRows := make([]map[string]any, len(ds.Mentors))
	for i, m := range ds.Mentors {
		mRows[i] = map[string]any{"mentor": m.MentorID, "mentee": m.MenteeID}
	}
	if err := store.Exec(ctx, `
		UNWIND $rows AS row
		MATCH (a:Person {id: row.mentor}), (b:Person {id: row.mentee})
		MERGE (a)-[:MENTORS]->(b)`,
		map[string]any{"rows": mRows}); err != nil {
		return fmt.Errorf("mentors: %w", err)
	}

	log.Println("seed complete ✓")
	return nil
}
