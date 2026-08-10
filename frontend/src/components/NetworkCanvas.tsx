import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { GraphNode, GraphLink } from "../lib/api";
import { useTheme } from "../lib/theme";

// Hand-rolled force-directed layout on <canvas> — no chart library.
// ~66 nodes and ~117 links, so O(n²) repulsion per frame is comfortably
// within budget. People render as glowing dots (sized by degree), projects
// as small squares; hovering highlights a node's neighbourhood and clicking
// a person opens their profile.

interface SimNode extends GraphNode {
  x: number;
  y: number;
  vx: number;
  vy: number;
  degree: number;
}

const PALETTES = {
  dark: {
    person: "#4d9fff",
    project: "#ff8a4d",
    linkDim: "rgba(148, 178, 255, 0.10)",
    linkLit: "rgba(77, 159, 255, 0.55)",
    glowBlur: 10,
  },
  light: {
    person: "#1f6fe0",
    project: "#e56a2b",
    linkDim: "rgba(20, 50, 120, 0.13)",
    linkLit: "rgba(31, 111, 224, 0.6)",
    glowBlur: 5,
  },
};

export function NetworkCanvas({
  nodes,
  links,
  height = 420,
}: {
  nodes: GraphNode[];
  links: GraphLink[];
  height?: number;
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const wrapRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const { theme } = useTheme();
  const COLORS = PALETTES[theme];
  const [tip, setTip] = useState<{ x: number; y: number; node: SimNode } | null>(null);
  const hoverRef = useRef<SimNode | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    const wrap = wrapRef.current;
    if (!canvas || !wrap || nodes.length === 0) return;

    const dpr = window.devicePixelRatio || 1;
    let width = wrap.clientWidth;
    canvas.width = width * dpr;
    canvas.height = height * dpr;
    canvas.style.height = `${height}px`;
    const ctx = canvas.getContext("2d")!;

    // deterministic starting positions on a ring, seeded by index
    const degree = new Map<string, number>();
    for (const l of links) {
      degree.set(l.source, (degree.get(l.source) ?? 0) + 1);
      degree.set(l.target, (degree.get(l.target) ?? 0) + 1);
    }
    const sim: SimNode[] = nodes.map((n, i) => {
      const angle = (i / nodes.length) * Math.PI * 2;
      const r = Math.min(width, height) * 0.36;
      return {
        ...n,
        x: width / 2 + Math.cos(angle) * r,
        y: height / 2 + Math.sin(angle) * r,
        vx: 0,
        vy: 0,
        degree: degree.get(n.id) ?? 0,
      };
    });
    const byId = new Map(sim.map((n) => [n.id, n]));
    const simLinks = links
      .map((l) => ({ a: byId.get(l.source)!, b: byId.get(l.target)! }))
      .filter((l) => l.a && l.b);
    const neighbours = new Map<string, Set<string>>();
    for (const { a, b } of simLinks) {
      (neighbours.get(a.id) ?? neighbours.set(a.id, new Set()).get(a.id)!).add(b.id);
      (neighbours.get(b.id) ?? neighbours.set(b.id, new Set()).get(b.id)!).add(a.id);
    }

    let raf = 0;
    let alpha = 1;

    const step = () => {
      // forces
      if (alpha > 0.005) {
        for (let i = 0; i < sim.length; i++) {
          const a = sim[i];
          for (let j = i + 1; j < sim.length; j++) {
            const b = sim[j];
            let dx = a.x - b.x;
            let dy = a.y - b.y;
            const d2 = Math.max(dx * dx + dy * dy, 40);
            const f = (900 / d2) * alpha;
            const d = Math.sqrt(d2);
            dx /= d;
            dy /= d;
            a.vx += dx * f;
            a.vy += dy * f;
            b.vx -= dx * f;
            b.vy -= dy * f;
          }
          // gentle gravity to centre
          a.vx += (width / 2 - a.x) * 0.0012 * alpha;
          a.vy += (height / 2 - a.y) * 0.0012 * alpha;
        }
        for (const { a, b } of simLinks) {
          const dx = b.x - a.x;
          const dy = b.y - a.y;
          const d = Math.max(Math.sqrt(dx * dx + dy * dy), 1);
          const f = ((d - 46) / d) * 0.02 * alpha;
          a.vx += dx * f;
          a.vy += dy * f;
          b.vx -= dx * f;
          b.vy -= dy * f;
        }
        for (const n of sim) {
          n.vx *= 0.85;
          n.vy *= 0.85;
          n.x = Math.max(14, Math.min(width - 14, n.x + n.vx));
          n.y = Math.max(14, Math.min(height - 14, n.y + n.vy));
        }
        alpha *= 0.995;
      }

      // draw
      const hover = hoverRef.current;
      const lit = hover
        ? new Set([hover.id, ...(neighbours.get(hover.id) ?? [])])
        : null;

      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      ctx.clearRect(0, 0, width, height);

      for (const { a, b } of simLinks) {
        const on = lit && (a.id === hover!.id || b.id === hover!.id);
        ctx.strokeStyle = on ? COLORS.linkLit : COLORS.linkDim;
        ctx.lineWidth = on ? 1.4 : 0.7;
        ctx.beginPath();
        ctx.moveTo(a.x, a.y);
        ctx.lineTo(b.x, b.y);
        ctx.stroke();
      }

      for (const n of sim) {
        const dimmed = lit !== null && !lit.has(n.id);
        const r = n.type === "person" ? 3.5 + Math.min(n.degree, 8) * 0.55 : 4;
        ctx.globalAlpha = dimmed ? 0.18 : 1;
        ctx.fillStyle = COLORS[n.type as "person" | "project"];
        ctx.shadowColor = ctx.fillStyle;
        ctx.shadowBlur = dimmed ? 0 : COLORS.glowBlur;
        ctx.beginPath();
        if (n.type === "person") {
          ctx.arc(n.x, n.y, r, 0, Math.PI * 2);
        } else {
          ctx.rect(n.x - r, n.y - r, r * 2, r * 2);
        }
        ctx.fill();
        ctx.shadowBlur = 0;
      }
      ctx.globalAlpha = 1;

      raf = requestAnimationFrame(step);
    };
    raf = requestAnimationFrame(step);

    const pick = (e: MouseEvent): SimNode | null => {
      const rect = canvas.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;
      let best: SimNode | null = null;
      let bestD = 14 * 14; // generous hit target
      for (const n of sim) {
        const dx = n.x - x;
        const dy = n.y - y;
        const d = dx * dx + dy * dy;
        if (d < bestD) {
          bestD = d;
          best = n;
        }
      }
      return best;
    };

    const onMove = (e: MouseEvent) => {
      const n = pick(e);
      hoverRef.current = n;
      setTip(n ? { x: n.x, y: n.y, node: n } : null);
      canvas.style.cursor = n?.type === "person" ? "pointer" : "crosshair";
    };
    const onLeave = () => {
      hoverRef.current = null;
      setTip(null);
    };
    const onClick = (e: MouseEvent) => {
      const n = pick(e);
      if (n?.type === "person") navigate(`/people/${n.id}`);
    };
    const onResize = () => {
      width = wrap.clientWidth;
      canvas.width = width * dpr;
      alpha = Math.max(alpha, 0.3);
    };

    canvas.addEventListener("mousemove", onMove);
    canvas.addEventListener("mouseleave", onLeave);
    canvas.addEventListener("click", onClick);
    window.addEventListener("resize", onResize);
    return () => {
      cancelAnimationFrame(raf);
      canvas.removeEventListener("mousemove", onMove);
      canvas.removeEventListener("mouseleave", onLeave);
      canvas.removeEventListener("click", onClick);
      window.removeEventListener("resize", onResize);
    };
  }, [nodes, links, height, navigate, theme]); // theme rebuild restarts the sim — an intentional flourish

  return (
    <div className="net-wrap" ref={wrapRef}>
      <canvas ref={canvasRef} aria-label="Collaboration network visualisation" />
      {tip && (
        <div className="net-tip" style={{ left: tip.x, top: tip.y }}>
          <strong>{tip.node.name}</strong>
          <div className="sub">
            {tip.node.type === "person"
              ? `${tip.node.title ?? ""} · ${tip.node.degree} project${tip.node.degree === 1 ? "" : "s"} — click to open`
              : "project"}
          </div>
        </div>
      )}
      <div className="net-legend">
        <span>
          <i style={{ background: COLORS.person, boxShadow: `0 0 8px ${COLORS.person}` }} />
          person
        </span>
        <span>
          <i style={{ background: COLORS.project, borderRadius: 2, boxShadow: `0 0 8px ${COLORS.project}` }} />
          project
        </span>
      </div>
    </div>
  );
}
