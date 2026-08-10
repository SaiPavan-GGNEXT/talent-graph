package seed

import (
	"math/rand"
	"strings"
)

// The dataset is generated deterministically (fixed rand seed) so every
// load produces the same graph — useful when walking through query results.

type Skill struct {
	Name     string
	Category string
}

type Client struct {
	Name     string
	Industry string
}

type Project struct {
	ID       string
	Name     string
	Client   string
	Year     int
	Status   string
	Requires []string
}

type Person struct {
	ID        string
	Name      string
	Title     string
	Seniority string
	Location  string
	Track     string
}

type HasSkill struct {
	PersonID string
	Skill    string
	Level    int
	Years    int
}

type WorkedOn struct {
	PersonID  string
	ProjectID string
	Role      string
}

type Mentors struct {
	MentorID string
	MenteeID string
}

type Dataset struct {
	Skills   []Skill
	Clients  []Client
	Projects []Project
	People   []Person
	HasSkill []HasSkill
	WorkedOn []WorkedOn
	Mentors  []Mentors
}

var skills = []Skill{
	{"Go", "Languages"}, {"Python", "Languages"}, {"TypeScript", "Languages"},
	{"Java", "Languages"}, {"Rust", "Languages"},
	{"React", "Frontend"}, {"Next.js", "Frontend"}, {"Design Systems", "Frontend"},
	{"Accessibility", "Frontend"},
	{"PostgreSQL", "Data"}, {"Cypher & Graph Modeling", "Data"}, {"Kafka", "Data"},
	{"Spark", "Data"}, {"dbt", "Data"},
	{"AWS", "Cloud"}, {"GCP", "Cloud"}, {"Kubernetes", "Cloud"},
	{"Terraform", "Cloud"}, {"Observability", "Cloud"},
	{"PyTorch", "AI/ML"}, {"LLM Integration", "AI/ML"}, {"MLOps", "AI/ML"},
	{"Computer Vision", "AI/ML"},
	{"Product Strategy", "Practice"}, {"UX Research", "Practice"},
	{"Agile Delivery", "Practice"}, {"Security Engineering", "Practice"},
}

var clients = []Client{
	{"Northwind Health", "Healthcare"},
	{"Aurora Bank", "Financial Services"},
	{"Vector Logistics", "Logistics"},
	{"Solstice Energy", "Energy"},
	{"Bluepeak Retail", "Retail"},
	{"Helio Media", "Media"},
	{"Cascade Insurance", "Insurance"},
	{"Trailhead Travel", "Travel"},
	{"Quanta Manufacturing", "Manufacturing"},
	{"Civic Labs", "Government"},
}

var projects = []Project{
	{"p-atlas", "Atlas — Patient Records Platform", "Northwind Health", 2023, "delivered", []string{"Go", "PostgreSQL", "React", "Security Engineering"}},
	{"p-ledger", "Ledger — Real-time Fraud Detection", "Aurora Bank", 2024, "delivered", []string{"Python", "Kafka", "PyTorch", "Observability"}},
	{"p-compass", "Compass — Fleet Route Optimizer", "Vector Logistics", 2024, "delivered", []string{"Go", "GCP", "Kubernetes", "PostgreSQL"}},
	{"p-flare", "Flare — Grid Anomaly Monitoring", "Solstice Energy", 2023, "delivered", []string{"Python", "Spark", "Computer Vision", "AWS"}},
	{"p-storefront", "Storefront — Headless Commerce", "Bluepeak Retail", 2024, "delivered", []string{"TypeScript", "Next.js", "Design Systems", "PostgreSQL"}},
	{"p-reel", "Reel — Recommendation Engine", "Helio Media", 2023, "delivered", []string{"Python", "PyTorch", "MLOps", "Kafka"}},
	{"p-shield", "Shield — Claims Automation", "Cascade Insurance", 2024, "delivered", []string{"Java", "LLM Integration", "PostgreSQL", "Agile Delivery"}},
	{"p-wander", "Wander — Trip Planning Copilot", "Trailhead Travel", 2025, "active", []string{"TypeScript", "LLM Integration", "React", "UX Research"}},
	{"p-forge", "Forge — Factory Telemetry Mesh", "Quanta Manufacturing", 2024, "delivered", []string{"Rust", "Kafka", "Kubernetes", "Observability"}},
	{"p-civic", "CivicOne — Permit Portal", "Civic Labs", 2023, "delivered", []string{"TypeScript", "React", "Accessibility", "PostgreSQL"}},
	{"p-pulse", "Pulse — Clinician Scheduling", "Northwind Health", 2025, "active", []string{"Go", "React", "PostgreSQL", "Agile Delivery"}},
	{"p-vault", "Vault — Open Banking APIs", "Aurora Bank", 2025, "active", []string{"Go", "Security Engineering", "AWS", "Terraform"}},
	{"p-relay", "Relay — Last-mile Tracking", "Vector Logistics", 2025, "active", []string{"TypeScript", "React", "Kafka", "GCP"}},
	{"p-helios", "Helios — Solar Yield Forecasting", "Solstice Energy", 2025, "active", []string{"Python", "PyTorch", "dbt", "MLOps"}},
	{"p-basket", "Basket — Personalization Engine", "Bluepeak Retail", 2025, "active", []string{"Python", "LLM Integration", "Kafka", "GCP"}},
	{"p-signal", "Signal — Newsroom Analytics", "Helio Media", 2025, "active", []string{"TypeScript", "dbt", "PostgreSQL", "Design Systems"}},
	{"p-anchor", "Anchor — Underwriting Graph", "Cascade Insurance", 2025, "active", []string{"Cypher & Graph Modeling", "Go", "LLM Integration", "PostgreSQL"}},
	{"p-summit", "Summit — Loyalty Replatform", "Trailhead Travel", 2024, "delivered", []string{"Java", "AWS", "PostgreSQL", "Agile Delivery"}},
}

// track -> core skills drawn from; titles per seniority
var tracks = map[string][]string{
	"backend":  {"Go", "Java", "Rust", "PostgreSQL", "Kafka", "Security Engineering"},
	"frontend": {"TypeScript", "React", "Next.js", "Design Systems", "Accessibility"},
	"data":     {"Python", "Spark", "dbt", "PostgreSQL", "Kafka", "Cypher & Graph Modeling"},
	"ml":       {"Python", "PyTorch", "LLM Integration", "MLOps", "Computer Vision"},
	"platform": {"Kubernetes", "Terraform", "AWS", "GCP", "Observability", "Go"},
	"product":  {"Product Strategy", "UX Research", "Agile Delivery"},
}

var trackTitles = map[string]string{
	"backend":  "Backend Engineer",
	"frontend": "Frontend Engineer",
	"data":     "Data Engineer",
	"ml":       "ML Engineer",
	"platform": "Platform Engineer",
	"product":  "Product Manager",
}

var names = []string{
	"Aarav Mehta", "Beatriz Costa", "Chen Wei", "Divya Raman", "Elena Petrova",
	"Farid Hassan", "Grace Okafor", "Hiroshi Tanaka", "Ines Fischer", "Jonas Lindqvist",
	"Kavya Nair", "Liam O'Sullivan", "Mariana Silva", "Noah Kimani", "Olivia Bennett",
	"Priya Sharma", "Quentin Moreau", "Rohan Gupta", "Sofia Rossi", "Tomas Novak",
	"Uma Krishnan", "Viktor Halasz", "Wanjiru Kamau", "Xavier Dubois", "Yara Haddad",
	"Zainab Ali", "Andre Johnson", "Bianca Torres", "Carlos Mendez", "Deepak Iyer",
	"Emma Walsh", "Felix Wagner", "Gita Patel", "Henrik Larsen", "Isabella Romano",
	"Jamal Carter", "Keiko Sato", "Lucas Almeida", "Meera Pillai", "Nina Kowalski",
	"Omar Farouk", "Paulo Ribeiro", "Rachel Kim", "Samuel Adeyemi", "Tara Fitzgerald",
	"Usha Reddy", "Vikram Singh", "Willem de Vries",
}

var locations = []string{"Hyderabad", "Bengaluru", "London", "Berlin", "New York", "Singapore", "Remote"}
var seniorities = []string{"Junior", "Mid", "Senior", "Staff", "Principal"}

// Build generates the full deterministic dataset.
func Build() Dataset {
	rng := rand.New(rand.NewSource(42))
	trackKeys := []string{"backend", "frontend", "data", "ml", "platform", "product"}

	ds := Dataset{Skills: skills, Clients: clients, Projects: projects}

	skillsByName := map[string]bool{}
	for _, s := range skills {
		skillsByName[s.Name] = true
	}

	// People with a specialty track and seniority.
	for i, name := range names {
		track := trackKeys[i%len(trackKeys)]
		sen := seniorities[rng.Intn(len(seniorities))]
		title := trackTitles[track]
		if sen == "Staff" || sen == "Principal" {
			title = sen + " " + title
		} else if sen == "Senior" {
			title = "Senior " + title
		}
		ds.People = append(ds.People, Person{
			ID:        slug(name),
			Name:      name,
			Title:     title,
			Seniority: sen,
			Location:  locations[rng.Intn(len(locations))],
			Track:     track,
		})
	}

	// Skills: most of the track's core skills plus 1-2 from elsewhere.
	allSkillNames := make([]string, len(skills))
	for i, s := range skills {
		allSkillNames[i] = s.Name
	}
	for _, p := range ds.People {
		core := tracks[p.Track]
		picked := map[string]bool{}
		n := 3 + rng.Intn(len(core)-2) // 3..len(core)
		perm := rng.Perm(len(core))
		for _, idx := range perm[:n] {
			picked[core[idx]] = true
		}
		for extras := rng.Intn(3); extras > 0; extras-- {
			picked[allSkillNames[rng.Intn(len(allSkillNames))]] = true
		}
		for name := range picked {
			level := 2 + rng.Intn(4) // 2..5
			if p.Seniority == "Junior" && level > 3 {
				level = 3
			}
			ds.HasSkill = append(ds.HasSkill, HasSkill{
				PersonID: p.ID, Skill: name,
				Level: level, Years: 1 + rng.Intn(9),
			})
		}
	}

	// Index: who has which skill, for staffing projects sensibly.
	bySkill := map[string][]string{}
	for _, hs := range ds.HasSkill {
		bySkill[hs.Skill] = append(bySkill[hs.Skill], hs.PersonID)
	}

	// Project teams: for each required skill pick 1-2 matching people,
	// then add a PM-track person to most projects.
	roles := []string{"Tech Lead", "Engineer", "Engineer", "Engineer", "Architect"}
	for _, pr := range ds.Projects {
		team := map[string]bool{}
		for _, req := range pr.Requires {
			candidates := bySkill[req]
			if len(candidates) == 0 {
				continue
			}
			for k := 0; k < 1+rng.Intn(2); k++ {
				team[candidates[rng.Intn(len(candidates))]] = true
			}
		}
		if rng.Float64() < 0.8 {
			pms := bySkill["Agile Delivery"]
			if len(pms) > 0 {
				team[pms[rng.Intn(len(pms))]] = true
			}
		}
		first := true
		for id := range team {
			role := roles[rng.Intn(len(roles))]
			if first {
				role = "Tech Lead"
				first = false
			}
			ds.WorkedOn = append(ds.WorkedOn, WorkedOn{PersonID: id, ProjectID: pr.ID, Role: role})
		}
	}

	// Mentoring: seniors/staff/principals mentor juniors and mids on their track.
	var seniorsBy, juniorsBy = map[string][]string{}, map[string][]string{}
	for _, p := range ds.People {
		switch p.Seniority {
		case "Senior", "Staff", "Principal":
			seniorsBy[p.Track] = append(seniorsBy[p.Track], p.ID)
		case "Junior", "Mid":
			juniorsBy[p.Track] = append(juniorsBy[p.Track], p.ID)
		}
	}
	for track, mentees := range juniorsBy {
		mentors := seniorsBy[track]
		if len(mentors) == 0 {
			continue
		}
		for i, mentee := range mentees {
			ds.Mentors = append(ds.Mentors, Mentors{
				MentorID: mentors[i%len(mentors)],
				MenteeID: mentee,
			})
		}
	}

	return ds
}

func slug(name string) string {
	s := strings.ToLower(name)
	return strings.NewReplacer("'", "", " ", "-").Replace(s)
}
