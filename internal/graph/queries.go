package graph

import (
	"context"
	"sort"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

// ---- API shapes ------------------------------------------------------------

type Stats struct {
	People   int64 `json:"people"`
	Skills   int64 `json:"skills"`
	Projects int64 `json:"projects"`
	Clients  int64 `json:"clients"`
	Rels     int64 `json:"relationships"`
}

type PersonSummary struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Title     string   `json:"title"`
	Seniority string   `json:"seniority"`
	Location  string   `json:"location"`
	TopSkills []string `json:"topSkills"`
}

type SkillLevel struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Level    int64  `json:"level"`
	Years    int64  `json:"years"`
}

type ProjectRef struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Year   int64  `json:"year"`
	Status string `json:"status"`
	Client string `json:"client"`
	Role   string `json:"role"`
}

type Colleague struct {
	Person         PersonSummary `json:"person"`
	SharedProjects int64         `json:"sharedProjects"`
}

type PersonDetail struct {
	PersonSummary
	Skills     []SkillLevel `json:"skills"`
	Projects   []ProjectRef `json:"projects"`
	Colleagues []Colleague  `json:"colleagues"`
	Mentor     *PersonSummary `json:"mentor,omitempty"`
	Mentees    []PersonSummary `json:"mentees"`
}

type SkillInfo struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	People   int64  `json:"people"`
}

// PathNode is one element of an introduction path. Type is "person" or
// "project"; Via describes the relationship that led here.
type PathNode struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
	Via   string `json:"via,omitempty"`
}

type Expert struct {
	Person   PersonSummary `json:"person"`
	Level    int64         `json:"level"`
	Years    int64         `json:"years"`
	Distance int64         `json:"distance"` // hops from the asking person; 0 = no origin given, -1 = unreachable
}

type AdjacentSkill struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Overlap  int64  `json:"overlap"` // people who hold both skills
}

type TeamMember struct {
	Person PersonSummary `json:"person"`
	Covers []SkillLevel  `json:"covers"`
}

type TeamPlan struct {
	Required  []string     `json:"required"`
	Team      []TeamMember `json:"team"`
	Uncovered []string     `json:"uncovered"`
}

// ---- Queries ---------------------------------------------------------------

// Stats returns headline counts for the dashboard.
func (s *Store) Stats(ctx context.Context) (Stats, error) {
	records, err := s.read(ctx, `
		MATCH (p:Person) WITH count(p) AS people
		MATCH (sk:Skill) WITH people, count(sk) AS skills
		MATCH (pr:Project) WITH people, skills, count(pr) AS projects
		MATCH (c:Client) WITH people, skills, projects, count(c) AS clients
		MATCH ()-[r]->() RETURN people, skills, projects, clients, count(r) AS rels`,
		nil)
	if err != nil {
		return Stats{}, err
	}
	if len(records) == 0 {
		return Stats{}, nil
	}
	rec := records[0]
	return Stats{
		People:   getInt(rec, "people"),
		Skills:   getInt(rec, "skills"),
		Projects: getInt(rec, "projects"),
		Clients:  getInt(rec, "clients"),
		Rels:     getInt(rec, "rels"),
	}, nil
}

// SearchPeople finds people by (case-insensitive) name or title fragment.
func (s *Store) SearchPeople(ctx context.Context, q string) ([]PersonSummary, error) {
	records, err := s.read(ctx, `
		MATCH (p:Person)
		WHERE toLower(p.name) CONTAINS toLower($q) OR toLower(p.title) CONTAINS toLower($q)
		OPTIONAL MATCH (p)-[r:HAS_SKILL]->(sk:Skill)
		WITH p, sk, r ORDER BY r.level DESC, sk.name
		WITH p, collect(sk.name)[..3] AS topSkills
		RETURN p, topSkills
		ORDER BY p.name LIMIT 25`,
		map[string]any{"q": q})
	if err != nil {
		return nil, err
	}
	out := make([]PersonSummary, 0, len(records))
	for _, rec := range records {
		out = append(out, personSummary(rec, "p", "topSkills"))
	}
	return out, nil
}

// GetPerson returns a person's full profile: skills, project history with
// clients, most frequent collaborators, and mentoring relationships.
func (s *Store) GetPerson(ctx context.Context, id string) (*PersonDetail, error) {
	records, err := s.read(ctx, `
		MATCH (p:Person {id: $id})
		OPTIONAL MATCH (p)-[hs:HAS_SKILL]->(sk:Skill)
		WITH p, collect({name: sk.name, category: sk.category, level: hs.level, years: hs.years}) AS skills
		OPTIONAL MATCH (p)-[w:WORKED_ON]->(pr:Project)-[:FOR_CLIENT]->(c:Client)
		WITH p, skills,
		     collect({id: pr.id, name: pr.name, year: pr.year, status: pr.status, client: c.name, role: w.role}) AS projects
		OPTIONAL MATCH (p)-[:WORKED_ON]->(:Project)<-[:WORKED_ON]-(col:Person)
		WITH p, skills, projects, col, count(*) AS shared
		ORDER BY shared DESC, col.name
		WITH p, skills, projects,
		     collect({person: col, shared: shared})[..6] AS colleagues
		OPTIONAL MATCH (mentor:Person)-[:MENTORS]->(p)
		OPTIONAL MATCH (p)-[:MENTORS]->(mentee:Person)
		RETURN p, skills, projects, colleagues, mentor, collect(DISTINCT mentee) AS mentees`,
		map[string]any{"id": id})
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	rec := records[0]

	detail := &PersonDetail{
		PersonSummary: personSummary(rec, "p", ""),
		Skills:        []SkillLevel{},
		Projects:      []ProjectRef{},
		Colleagues:    []Colleague{},
		Mentees:       []PersonSummary{},
	}

	for _, item := range getList(rec, "skills") {
		m, ok := item.(map[string]any)
		if !ok || m["name"] == nil {
			continue
		}
		detail.Skills = append(detail.Skills, SkillLevel{
			Name:     str(m["name"]),
			Category: str(m["category"]),
			Level:    i64(m["level"]),
			Years:    i64(m["years"]),
		})
	}
	sort.Slice(detail.Skills, func(a, b int) bool {
		if detail.Skills[a].Level != detail.Skills[b].Level {
			return detail.Skills[a].Level > detail.Skills[b].Level
		}
		return detail.Skills[a].Name < detail.Skills[b].Name
	})

	for _, item := range getList(rec, "projects") {
		m, ok := item.(map[string]any)
		if !ok || m["id"] == nil {
			continue
		}
		detail.Projects = append(detail.Projects, ProjectRef{
			ID:     str(m["id"]),
			Name:   str(m["name"]),
			Year:   i64(m["year"]),
			Status: str(m["status"]),
			Client: str(m["client"]),
			Role:   str(m["role"]),
		})
	}
	sort.Slice(detail.Projects, func(a, b int) bool {
		return detail.Projects[a].Year > detail.Projects[b].Year
	})

	for _, item := range getList(rec, "colleagues") {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		node, ok := m["person"].(dbtype.Node)
		if !ok {
			continue
		}
		detail.Colleagues = append(detail.Colleagues, Colleague{
			Person:         nodeToPerson(node),
			SharedProjects: i64(m["shared"]),
		})
	}

	if mentorVal, ok := rec.Get("mentor"); ok {
		if node, ok := mentorVal.(dbtype.Node); ok {
			mentor := nodeToPerson(node)
			detail.Mentor = &mentor
		}
	}
	for _, item := range getList(rec, "mentees") {
		if node, ok := item.(dbtype.Node); ok {
			detail.Mentees = append(detail.Mentees, nodeToPerson(node))
		}
	}
	return detail, nil
}

// ListSkills returns every skill with how many people hold it.
func (s *Store) ListSkills(ctx context.Context) ([]SkillInfo, error) {
	records, err := s.read(ctx, `
		MATCH (sk:Skill)
		OPTIONAL MATCH (sk)<-[:HAS_SKILL]-(p:Person)
		RETURN sk.name AS name, sk.category AS category, count(p) AS people
		ORDER BY category, name`,
		nil)
	if err != nil {
		return nil, err
	}
	out := make([]SkillInfo, 0, len(records))
	for _, rec := range records {
		out = append(out, SkillInfo{
			Name:     getStr(rec, "name"),
			Category: getStr(rec, "category"),
			People:   getInt(rec, "people"),
		})
	}
	return out, nil
}

// IntroPath finds the shortest chain of shared projects and mentoring links
// connecting two people — the "warm introduction" path. Variable-length
// shortest path is the query a relational database finds genuinely awkward:
// in SQL this is a recursive CTE with cycle detection and manual best-path
// pruning; in Cypher it is one MATCH.
func (s *Store) IntroPath(ctx context.Context, fromID, toID string) ([]PathNode, error) {
	records, err := s.read(ctx, `
		MATCH (a:Person {id: $from}), (b:Person {id: $to})
		MATCH path = shortestPath((a)-[:WORKED_ON|MENTORS*..8]-(b))
		RETURN path`,
		map[string]any{"from": fromID, "to": toID})
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	pathVal, ok := records[0].Get("path")
	if !ok {
		return nil, nil
	}
	path, ok := pathVal.(dbtype.Path)
	if !ok {
		return nil, nil
	}

	nodes := make([]PathNode, 0, len(path.Nodes))
	for i, n := range path.Nodes {
		pn := PathNode{ID: str(n.Props["id"]), Name: str(n.Props["name"])}
		if hasLabel(n, "Person") {
			pn.Type = "person"
			pn.Title = str(n.Props["title"])
		} else {
			pn.Type = "project"
		}
		if i > 0 && i-1 < len(path.Relationships) {
			pn.Via = path.Relationships[i-1].Type
		}
		nodes = append(nodes, pn)
	}
	return nodes, nil
}

// Experts finds people who hold a skill, ranked by proficiency. When fromID
// is given, each expert is annotated with their collaboration distance from
// that person (a 2+ hop traversal through shared projects), so results can
// be sorted by "who can I actually reach".
func (s *Store) Experts(ctx context.Context, skill, fromID string) ([]Expert, error) {
	records, err := s.read(ctx, `
		MATCH (sk:Skill {name: $skill})<-[hs:HAS_SKILL]-(p:Person)
		WHERE $from = '' OR p.id <> $from
		OPTIONAL MATCH (p)-[r2:HAS_SKILL]->(other:Skill)
		WITH sk, hs, p, other, r2 ORDER BY r2.level DESC
		WITH sk, hs, p, collect(other.name)[..3] AS topSkills
		OPTIONAL MATCH (me:Person {id: $from})
		OPTIONAL MATCH dist = shortestPath((me)-[:WORKED_ON*..6]-(p))
		RETURN p, topSkills, hs.level AS level, hs.years AS years,
		       CASE
		         WHEN $from = '' THEN 0
		         WHEN dist IS NULL THEN -1
		         ELSE length(dist) / 2
		       END AS distance
		ORDER BY level DESC, years DESC
		LIMIT 30`,
		map[string]any{"skill": skill, "from": fromID})
	if err != nil {
		return nil, err
	}
	out := make([]Expert, 0, len(records))
	for _, rec := range records {
		out = append(out, Expert{
			Person:   personSummary(rec, "p", "topSkills"),
			Level:    getInt(rec, "level"),
			Years:    getInt(rec, "years"),
			Distance: getInt(rec, "distance"),
		})
	}
	return out, nil
}

// AdjacentSkills answers "people who know X usually also know…" —
// a co-occurrence traversal used for skill-gap and hiring planning.
func (s *Store) AdjacentSkills(ctx context.Context, skill string) ([]AdjacentSkill, error) {
	records, err := s.read(ctx, `
		MATCH (sk:Skill {name: $skill})<-[:HAS_SKILL]-(p:Person)-[:HAS_SKILL]->(other:Skill)
		WHERE other <> sk
		RETURN other.name AS name, other.category AS category, count(p) AS overlap
		ORDER BY overlap DESC
		LIMIT 10`,
		map[string]any{"skill": skill})
	if err != nil {
		return nil, err
	}
	out := make([]AdjacentSkill, 0, len(records))
	for _, rec := range records {
		out = append(out, AdjacentSkill{
			Name:     getStr(rec, "name"),
			Category: getStr(rec, "category"),
			Overlap:  getInt(rec, "overlap"),
		})
	}
	return out, nil
}

// TeamPlan suggests the smallest team covering the required skills.
// Cypher fetches every candidate per skill in one round-trip; the greedy
// set-cover happens here in Go where it is easy to read and test.
func (s *Store) TeamPlan(ctx context.Context, skills []string) (*TeamPlan, error) {
	records, err := s.read(ctx, `
		UNWIND $skills AS want
		MATCH (sk:Skill {name: want})<-[hs:HAS_SKILL]-(p:Person)
		RETURN want, p, hs.level AS level, hs.years AS years
		ORDER BY level DESC, years DESC`,
		map[string]any{"skills": skills})
	if err != nil {
		return nil, err
	}

	// candidate -> covered skills, and remember person nodes / proficiency
	type cover struct {
		person PersonSummary
		skills map[string]SkillLevel
		score  int64
	}
	candidates := map[string]*cover{}
	for _, rec := range records {
		nodeVal, _ := rec.Get("p")
		node, ok := nodeVal.(dbtype.Node)
		if !ok {
			continue
		}
		p := nodeToPerson(node)
		c, exists := candidates[p.ID]
		if !exists {
			c = &cover{person: p, skills: map[string]SkillLevel{}}
			candidates[p.ID] = c
		}
		name := getStr(rec, "want")
		lvl := getInt(rec, "level")
		c.skills[name] = SkillLevel{Name: name, Level: lvl, Years: getInt(rec, "years")}
		c.score += lvl
	}

	remaining := map[string]bool{}
	for _, s := range skills {
		remaining[s] = true
	}

	plan := &TeamPlan{Required: skills, Team: []TeamMember{}, Uncovered: []string{}}
	for len(remaining) > 0 && len(candidates) > 0 {
		var best *cover
		bestGain := 0
		var bestID string
		// deterministic iteration: sort candidate ids
		ids := make([]string, 0, len(candidates))
		for id := range candidates {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			c := candidates[id]
			gain := 0
			for name := range c.skills {
				if remaining[name] {
					gain++
				}
			}
			if gain > bestGain || (gain == bestGain && gain > 0 && best != nil && c.score > best.score) {
				best, bestGain, bestID = c, gain, id
			}
		}
		if best == nil || bestGain == 0 {
			break
		}
		member := TeamMember{Person: best.person, Covers: []SkillLevel{}}
		for name, sl := range best.skills {
			if remaining[name] {
				member.Covers = append(member.Covers, sl)
				delete(remaining, name)
			}
		}
		sort.Slice(member.Covers, func(a, b int) bool { return member.Covers[a].Name < member.Covers[b].Name })
		plan.Team = append(plan.Team, member)
		delete(candidates, bestID)
	}
	for name := range remaining {
		plan.Uncovered = append(plan.Uncovered, name)
	}
	sort.Strings(plan.Uncovered)
	return plan, nil
}

// ListPeople returns everyone, for pickers and the directory page.
func (s *Store) ListPeople(ctx context.Context) ([]PersonSummary, error) {
	records, err := s.read(ctx, `
		MATCH (p:Person)
		OPTIONAL MATCH (p)-[r:HAS_SKILL]->(sk:Skill)
		WITH p, sk, r ORDER BY r.level DESC, sk.name
		WITH p, collect(sk.name)[..3] AS topSkills
		RETURN p, topSkills ORDER BY p.name`,
		nil)
	if err != nil {
		return nil, err
	}
	out := make([]PersonSummary, 0, len(records))
	for _, rec := range records {
		out = append(out, personSummary(rec, "p", "topSkills"))
	}
	return out, nil
}

// ---- network view -------------------------------------------------------------

type GraphNode struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"` // "person" | "project"
	Title string `json:"title,omitempty"`
}

type GraphLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type GraphView struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

// NetworkView returns the whole collaboration fabric (people, projects and
// WORKED_ON edges) for the animated network visualisation. The dataset is
// deliberately small (free-tier sized), so one round-trip returns it all.
func (s *Store) GraphView(ctx context.Context) (*GraphView, error) {
	records, err := s.read(ctx, `
		MATCH (p:Person)
		OPTIONAL MATCH (p)-[:WORKED_ON]->(pr:Project)
		RETURN p.id AS pid, p.name AS pname, p.title AS ptitle,
		       pr.id AS prid, pr.name AS prname`,
		nil)
	if err != nil {
		return nil, err
	}
	view := &GraphView{Nodes: []GraphNode{}, Links: []GraphLink{}}
	seen := map[string]bool{}
	for _, rec := range records {
		pid := getStr(rec, "pid")
		if !seen[pid] {
			seen[pid] = true
			view.Nodes = append(view.Nodes, GraphNode{
				ID: pid, Name: getStr(rec, "pname"), Type: "person", Title: getStr(rec, "ptitle"),
			})
		}
		prid := getStr(rec, "prid")
		if prid == "" {
			continue
		}
		if !seen[prid] {
			seen[prid] = true
			view.Nodes = append(view.Nodes, GraphNode{
				ID: prid, Name: getStr(rec, "prname"), Type: "project",
			})
		}
		view.Links = append(view.Links, GraphLink{Source: pid, Target: prid})
	}
	return view, nil
}

// ---- record helpers ----------------------------------------------------------

func personSummary(rec *neo4j.Record, nodeKey, skillsKey string) PersonSummary {
	val, _ := rec.Get(nodeKey)
	node, ok := val.(dbtype.Node)
	if !ok {
		return PersonSummary{}
	}
	p := nodeToPerson(node)
	if skillsKey != "" {
		for _, s := range getList(rec, skillsKey) {
			if name := str(s); name != "" {
				p.TopSkills = append(p.TopSkills, name)
			}
		}
	}
	return p
}

func nodeToPerson(n dbtype.Node) PersonSummary {
	return PersonSummary{
		ID:        str(n.Props["id"]),
		Name:      str(n.Props["name"]),
		Title:     str(n.Props["title"]),
		Seniority: str(n.Props["seniority"]),
		Location:  str(n.Props["location"]),
		TopSkills: []string{},
	}
}

func hasLabel(n dbtype.Node, label string) bool {
	for _, l := range n.Labels {
		if l == label {
			return true
		}
	}
	return false
}

func getStr(rec *neo4j.Record, key string) string {
	v, _ := rec.Get(key)
	return str(v)
}

func getInt(rec *neo4j.Record, key string) int64 {
	v, _ := rec.Get(key)
	return i64(v)
}

func getList(rec *neo4j.Record, key string) []any {
	v, _ := rec.Get(key)
	list, _ := v.([]any)
	return list
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func i64(v any) int64 {
	n, _ := v.(int64)
	return n
}
