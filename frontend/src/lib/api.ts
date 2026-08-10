// Typed API client. All errors funnel into ApiError so pages can render a
// dedicated "database offline" state (HTTP 503) vs. a generic failure.

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public offline: boolean,
  ) {
    super(message);
  }
}

export interface Stats {
  people: number;
  skills: number;
  projects: number;
  clients: number;
  relationships: number;
}

export interface PersonSummary {
  id: string;
  name: string;
  title: string;
  seniority: string;
  location: string;
  topSkills: string[];
}

export interface SkillLevel {
  name: string;
  category: string;
  level: number;
  years: number;
}

export interface ProjectRef {
  id: string;
  name: string;
  year: number;
  status: string;
  client: string;
  role: string;
}

export interface Colleague {
  person: PersonSummary;
  sharedProjects: number;
}

export interface PersonDetail extends PersonSummary {
  skills: SkillLevel[];
  projects: ProjectRef[];
  colleagues: Colleague[];
  mentor?: PersonSummary;
  mentees: PersonSummary[];
}

export interface SkillInfo {
  name: string;
  category: string;
  people: number;
}

export interface PathNode {
  type: "person" | "project";
  id: string;
  name: string;
  title?: string;
  via?: string;
}

export interface PathResult {
  path: PathNode[] | null;
  connected: boolean;
}

export interface Expert {
  person: PersonSummary;
  level: number;
  years: number;
  distance: number; // 0 = no origin given, -1 = unreachable
}

export interface AdjacentSkill {
  name: string;
  category: string;
  overlap: number;
}

export interface TeamMember {
  person: PersonSummary;
  covers: SkillLevel[];
}

export interface GraphNode {
  id: string;
  name: string;
  type: "person" | "project";
  title?: string;
}

export interface GraphLink {
  source: string;
  target: string;
}

export interface GraphViewData {
  nodes: GraphNode[];
  links: GraphLink[];
}

export interface TeamPlan {
  required: string[];
  team: TeamMember[];
  uncovered: string[];
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(path, init);
  } catch {
    throw new ApiError("Could not reach the server.", 0, true);
  }
  if (!res.ok) {
    let message = `Request failed (${res.status})`;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
      // non-JSON error body; keep default message
    }
    throw new ApiError(message, res.status, res.status === 503);
  }
  return res.json();
}

export const api = {
  stats: () => request<Stats>("/api/stats"),
  graph: () => request<GraphViewData>("/api/graph"),
  people: (q?: string) =>
    request<PersonSummary[]>(`/api/people${q ? `?q=${encodeURIComponent(q)}` : ""}`),
  person: (id: string) => request<PersonDetail>(`/api/people/${encodeURIComponent(id)}`),
  skills: () => request<SkillInfo[]>("/api/skills"),
  experts: (skill: string, from?: string) =>
    request<Expert[]>(
      `/api/experts?skill=${encodeURIComponent(skill)}${from ? `&from=${encodeURIComponent(from)}` : ""}`,
    ),
  adjacentSkills: (skill: string) =>
    request<AdjacentSkill[]>(`/api/skills/adjacent?skill=${encodeURIComponent(skill)}`),
  path: (from: string, to: string) =>
    request<PathResult>(
      `/api/path?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
    ),
  teamPlan: (skills: string[]) =>
    request<TeamPlan>("/api/team-plan", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ skills }),
    }),
};
