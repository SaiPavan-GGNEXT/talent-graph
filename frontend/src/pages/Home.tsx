import { Link } from "react-router-dom";
import { api } from "../lib/api";
import { useApi } from "../lib/useApi";
import { StateGate } from "../components/StateGate";
import { NetworkCanvas } from "../components/NetworkCanvas";

const tools = [
  {
    to: "/experts",
    title: "Find experts",
    desc: "Who actually knows Kubernetes? Ranked by proficiency — and by how close they are to you in the collaboration network.",
  },
  {
    to: "/path",
    title: "Intro paths",
    desc: "The shortest chain of shared projects and mentors connecting any two people. Ask for a warm intro, not a cold ping.",
  },
  {
    to: "/team",
    title: "Team builder",
    desc: "Pick the skills a project needs and get the smallest team that covers them all.",
  },
  {
    to: "/people",
    title: "Directory",
    desc: "Browse everyone, their skills, project history, collaborators and mentoring lines.",
  },
];

export default function Home() {
  const stats = useApi(() => api.stats(), []);
  const graph = useApi(() => api.graph(), []);

  return (
    <>
      <div className="eyebrow" style={{ marginTop: 48 }}>
        Expertise · Collaboration · Graph
      </div>
      <h1 style={{ marginTop: 10 }}>
        Who knows what — <span className="gradient-text">and who can reach them.</span>
      </h1>
      <p className="lede">
        TalentGraph maps people, skills, projects and clients as a living network, so
        questions that span relationships — <em>“find me a Kafka expert two intros
        away”</em> — are one traversal, not a pile of joins.
      </p>

      <StateGate loading={stats.loading} error={stats.error} onRetry={stats.retry}>
        {stats.data && (
          <div className="stat-row">
            <div className="stat">
              <div className="value">{stats.data.people}</div>
              <div className="label">People</div>
            </div>
            <div className="stat">
              <div className="value">{stats.data.skills}</div>
              <div className="label">Skills</div>
            </div>
            <div className="stat">
              <div className="value">{stats.data.projects}</div>
              <div className="label">Projects</div>
            </div>
            <div className="stat">
              <div className="value">{stats.data.clients}</div>
              <div className="label">Clients</div>
            </div>
            <div className="stat">
              <div className="value">{stats.data.relationships}</div>
              <div className="label">Edges</div>
            </div>
          </div>
        )}
      </StateGate>

      <StateGate loading={graph.loading} error={graph.error} onRetry={graph.retry}>
        {graph.data && (
          <>
            <NetworkCanvas nodes={graph.data.nodes} links={graph.data.links} />
            <p className="section-note" style={{ textAlign: "center" }}>
              The live collaboration fabric — every person, project and engagement in
              the graph. Hover to inspect, click a person to open their profile.
            </p>
          </>
        )}
      </StateGate>

      <h2>Explore</h2>
      <div className="grid cols-2">
        {tools.map((t) => (
          <Link key={t.to} to={t.to} className="card person-card">
            <div className="person-name">{t.title}</div>
            <p style={{ margin: "6px 0 0", color: "var(--ink-2)" }}>{t.desc}</p>
          </Link>
        ))}
      </div>
    </>
  );
}
