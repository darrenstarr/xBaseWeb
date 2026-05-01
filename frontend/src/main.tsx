import React from "react";
import ReactDOM from "react-dom/client";

interface ScreenLine { row?: number; col?: number; text: string }
interface ScreenField { row?: number; col?: number; var: string; picture?: string; value?: string; type: string; options?: { value: string; label: string }[] }
interface TableColumn { name: string; align?: string }
interface RowAction { label: string; procedure: string }
interface TableData { columns: TableColumn[]; rows: string[][]; actions?: RowAction[]; keyCol?: number; query?: string; limit?: number; offset?: number; total?: number }
interface Screen { lines: ScreenLine[]; fields: ScreenField[]; prompt?: string; wait?: boolean; done?: boolean; result?: string; table?: TableData; title?: string; tagline?: string; nav?: Record<string, string> }

function applyMask(raw: string, mask: string): string {
    let ri = 0, out = "";
    for (const mc of mask) {
        if (ri >= raw.length) break;
        if ("9XxA!".includes(mc)) { out += raw[ri++]; }
        else { out += mc; }
    }
    return out;
}

// Clean input to only contain characters valid for the mask type,
// removing any literal characters from previous formatting.
function cleanForMask(raw: string, mask: string): string {
    if (mask.includes("9")) return raw.replace(/\D/g, "");

    // Build a set of allowed characters: mask format chars
    // X=any, A=alpha, !=uppercase. Everything else is a literal to strip.
    let out = "";
    for (const c of raw) {
        // Check if this character would be accepted at any position in the mask
        // Simple approach: keep anything that isn't a mask literal
        let keep = true;
        for (const mc of mask) {
            if (!"9XxA!".includes(mc) && c === mc) { keep = false; break; }
        }
        if (keep) out += c;
    }
    return out;
}

const MAIN_MENU_PROC = "MainMenu";
const MENU_MAP: Record<string, string> = {
  "1": "CustomerMenu", "2": "ApptMenu", "3": "ServicesMenu", "4": "InvoiceMenu", "0": "QUIT",
};
const SUB_MENU_MAP: Record<string, Record<string, string>> = {
  CustomerMenu: { "1": "AddCustomer", "2": "ListCustomers", "0": MAIN_MENU_PROC },
  ApptMenu: { "1": "AddAppointment", "2": "ListAppointments", "0": MAIN_MENU_PROC },
  ServicesMenu: { "1": "AddService", "2": "ListServices", "0": MAIN_MENU_PROC },
  InvoiceMenu: { "1": "ListInvoices", "2": "OverdueAccounts", "3": "GenerateInvoice", "0": MAIN_MENU_PROC },
};
const PARENT_MAP: Record<string, string> = {
  AddCustomer: "CustomerMenu", EditCustomer: "CustomerMenu",
  ListCustomers: "CustomerMenu", DeleteCustomer: "CustomerMenu",
  AddAppointment: "ApptMenu", EditAppointment: "ApptMenu",
  ListAppointments: "ApptMenu", CompleteAppt: "ApptMenu",
  CancelAppt: "ApptMenu", DeleteAppt: "ApptMenu",
  AddService: "ServicesMenu", EditService: "ServicesMenu",
  ListServices: "ServicesMenu", DeleteService: "ServicesMenu",
  ListInvoices: "InvoiceMenu", RecordPayment: "InvoiceMenu",
  DeleteInvoice: "InvoiceMenu", OverdueAccounts: "InvoiceMenu",
  GenerateInvoice: "InvoiceMenu",
};
const findParentMenu = (p: string): string => PARENT_MAP[p] || MAIN_MENU_PROC;

function App() {
  const [theme, setTheme] = React.useState<Record<string, string> | null>(null);
  const [screen, setScreen] = React.useState<Screen | null>(null);
  const [fieldVals, setFieldVals] = React.useState<Record<string, string>>({});
  const [loading, setLoading] = React.useState(false);
  const [currentProc, setCurrentProc] = React.useState(MAIN_MENU_PROC);
  const [navStack, setNavStack] = React.useState<string[]>([]);
  const [appTitle, setAppTitle] = React.useState("db4web");
  const [appTagline, setAppTagline] = React.useState("");

  const goBack = () => {
    if (navStack.length > 0) {
      const prev = navStack[navStack.length - 1];
      setNavStack(prev => prev.slice(0, -1));
      setCurrentProc(prev);
    }
  };

  React.useEffect(() => {
    fetch("/api/workspace").then(r => r.json()).then((c: any) => setTheme(c.theme)).catch(() => setTheme({
      background: "#0d1117", surface: "#161b22", text: "#c9d1d9", textMuted: "#8b949e",
      accent: "#58a6ff", accentGreen: "#3fb950", accentRed: "#da3633", border: "#30363d",
      font: "'Courier New', monospace",
    }));
    runInterpreter(MAIN_MENU_PROC, {});
  }, []);

  const runInterpreter = async (proc: string, input: Record<string, string>) => {
    setLoading(true); setCurrentProc(proc); setFieldVals({});
    try {
      const res = await fetch("/api/execute", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ program: "examples/app.prg", procedure: proc, state: { input } }),
      });
      const data = await res.json();
      const s = data.screen as Screen;
      setScreen(s);
      const init: Record<string, string> = {};
      s?.fields?.forEach((f: ScreenField) => { init[f.var] = f.value || ""; });
      if (Object.keys(init).length > 0) setFieldVals(init);
      if (s?.title) setAppTitle(s.title);
      if (s?.tagline) setAppTagline(s.tagline);
    } finally { setLoading(false); }
  };

  const handleRowAction = (action: RowAction, row: string[], keyCol: number) => {
    const key = row[keyCol] || "";
    runInterpreter(action.procedure, { mId: key });
  };

  const handleSubmit = async (directChoice?: string) => {
    if (!screen) return;
    if (screen.done) {
      runInterpreter(currentProc === MAIN_MENU_PROC ? MAIN_MENU_PROC : findParentMenu(currentProc), {});
      return;
    }
    const choice = directChoice ?? (screen.prompt ? (fieldVals[screen.prompt] || "") : "");

    // Use screen.nav for menu navigation (defined in .prg via NAV statement)
    if (screen.nav && screen.nav[choice]) {
      const target = screen.nav[choice];
      if (target === "QUIT" || target === "BACK") {
        setScreen({ lines: [{ row: 1, col: 1, text: target === "QUIT" ? "Goodbye!" : "" }], fields: [], done: true });
        return;
      }
      runInterpreter(target, {});
      return;
    }

    // Check hardcoded maps as fallback (can remove once all .prg have NAV)
    const sub = SUB_MENU_MAP[currentProc];
    if (sub && sub[choice]) {
      runInterpreter(sub[choice] === MAIN_MENU_PROC ? MAIN_MENU_PROC : sub[choice], {});
      return;
    }
    const target = MENU_MAP[choice];
    if (target === "QUIT") {
      setScreen({ lines: [{ row: 1, col: 1, text: "Goodbye!" }], fields: [], done: true });
      return;
    }
    if (target && currentProc === MAIN_MENU_PROC) {
      runInterpreter(target, {});
      return;
    }
    // Everything else: send current field values back to the interpreter
    runInterpreter(currentProc, fieldVals);
  };

  const t = theme || {};
  return (
    <div style={{ minHeight: "100vh", display: "flex", flexDirection: "column", backgroundColor: t.background || "#0d1117", color: t.text || "#c9d1d9", fontFamily: t.font || "'Courier New', monospace" }}>
      <header style={{ background: t.surface || "#161b22", borderBottom: `2px solid ${t.accent || "#58a6ff"}`, padding: "0 24px" }}>
        <div onClick={() => runInterpreter(MAIN_MENU_PROC, {})} style={{ display: "flex", alignItems: "center", gap: 12, padding: "12px 0", cursor: "pointer" }}>
          <span style={{ fontSize: 28, color: t.accent }}>&#9876;</span>
          <div>
            <h1 style={{ margin: 0, fontSize: 20, fontWeight: "bold", color: t.text, letterSpacing: 1 }}>{appTitle}</h1>
            {appTagline && <p style={{ margin: 0, fontSize: 12, color: t.accent, fontStyle: "italic" }}>{appTagline}</p>}
          </div>
        </div>
      </header>

      {navStack.length > 0 && (
        <div style={{ backgroundColor: t.surface || "#161b22", borderBottom: `1px solid ${t.border || "#30363d"}`, padding: "4px 24px" }}>
          <button onClick={goBack} style={{
            background: "none", border: `1px solid ${t.border || "#30363d"}`, borderRadius: 4, color: t.accent || "#58a6ff",
            fontFamily: t.font, fontSize: 12, cursor: "pointer", padding: "4px 12px",
          }}>&larr; Back</button>
          <span style={{ marginLeft: 12, fontSize: 11, color: t.textMuted || "#8b949e" }}>
            {navStack.map((v, i) => <span key={i}>{i > 0 && " / "}{v}</span>)}
          </span>
        </div>
      )}

      <main style={{ flex: 1, padding: 24, maxWidth: 900, width: "100%", margin: "0 auto", boxSizing: "border-box" }}>
        <div style={{ marginBottom: 16 }}>
          <h2 style={{ margin: 0, fontSize: 18, color: t.text || "#c9d1d9", fontWeight: "bold" }}>
            {currentProc.replace(/([A-Z])/g, ' $1').replace(/^(.)/, c => c.toUpperCase()).trim()}
          </h2>
          {loading && <span style={{ fontSize: 11, color: t.accent }}> executing...</span>}
        </div>

        <div style={{ backgroundColor: "#0a0a0f", border: `1px solid ${t.border || "#30363d"}`, borderRadius: 8, padding: 24, fontSize: 14, lineHeight: 1.6 }}>
          {/* Non-table lines (menu, headers, labels) */}
          {(screen?.lines || []).map((l, i) => {
            const isMenuItem = !!l.text.match(/^\s*\d+\.\s/);
            const isHeader = l.text.includes("===") && !l.text.includes("---");
            return (
              <div key={i} onClick={() => {
                if (isMenuItem && screen?.fields?.length) {
                  const m = l.text.match(/^\s*(\d+)/);
                  if (m) handleSubmit(m[1]);
                }
              }}
                style={{
                  color: isHeader ? t.accent : isMenuItem ? t.accent : t.textMuted,
                  fontWeight: isHeader ? "bold" : "normal",
                  cursor: isMenuItem ? "pointer" : "default",
                  padding: isMenuItem ? "2px 4px" : 0, borderRadius: 4,
                  transition: "background 0.1s",
                }}
                onMouseEnter={e => { if (isMenuItem) e.currentTarget.style.background = t.surface || "#161b22"; }}
                onMouseLeave={e => { if (isMenuItem) e.currentTarget.style.background = "transparent"; }}
              >{l.text}</div>
            );
          })}

          {/* Structured table rendering with sorting and infinite scroll */}
          {screen?.table && screen.table.columns.length > 0 && (
            <TableWithScroll table={screen.table} theme={t} onRowAction={handleRowAction} />
          )}

          {(screen?.fields || []).filter(() => {
            return !(screen?.lines?.some?.(l => /^\s*\d+\.\s/.test(l.text)) && screen.fields?.length === 1);
          }).map((f, i) => (
            <div key={i} style={{ marginTop: 8 }}>
              <div style={{ fontSize: 11, color: t.textMuted, marginBottom: 4 }}>{f.var.replace(/^m/, "")}</div>
              {f.options ? (
                <select value={fieldVals[f.var] ?? ""} onChange={e => setFieldVals({ ...fieldVals, [f.var]: e.target.value })}
                  style={{
                    backgroundColor: t.background, border: `1px solid ${t.border}`, borderRadius: 4,
                    padding: "10px 14px", fontFamily: t.font, fontSize: 16, color: t.text, outline: "none",
                    width: Math.max((f.picture ? f.picture.length * 10 : 300), 200),
                  }}>
                  <option value="">-- Select --</option>
                  {f.options.map((o, oi) => <option key={oi} value={o.value}>{o.label}</option>)}
                </select>
              ) : f.picture?.match(/9999-99-99 99:99/) ? (
                <input autoFocus={i === 0} type="datetime-local"
                  value={(fieldVals[f.var] ?? "").replace(/^NIL$/, "")}
                  onChange={e => setFieldVals({ ...fieldVals, [f.var]: e.target.value })}
                  style={{
                    backgroundColor: t.background, border: `1px solid ${t.border}`, borderRadius: 4,
                    padding: "10px 14px", fontFamily: t.font, fontSize: 16, color: t.text, outline: "none",
                    width: 250,
                  }}
                />
              ) : (
                <input autoFocus={i === 0}
                  value={(fieldVals[f.var] ?? "").replace(/^NIL$/, "")}
                  onChange={e => {
                    const raw = e.target.value;
                    if (f.picture) {
                      const clean = f.picture.includes("9") ? raw.replace(/\D/g, "") : cleanForMask(raw, f.picture);
                      setFieldVals({ ...fieldVals, [f.var]: applyMask(clean, f.picture) });
                    } else { setFieldVals({ ...fieldVals, [f.var]: raw }); }
                  }}
                  onKeyDown={e => { if (e.key === "Enter") handleSubmit(); }}
                  style={{
                    backgroundColor: t.background, border: `1px solid ${t.border}`, borderRadius: 4,
                    padding: "10px 14px", fontFamily: t.font, fontSize: 16, color: t.text, outline: "none",
                    width: Math.max((f.picture ? f.picture.length * 10 : 200), 150), letterSpacing: 2,
                  }}
                />
              )}
            </div>
          ))}

          {screen?.fields && screen.fields.length > 0 && !(screen.fields.length === 1 && (screen.lines || []).some(l => /^\s*\d+\.\s/.test(l.text))) && (
            <div style={{ marginTop: 12, display: "flex", gap: 8 }}>
              <button onClick={() => handleSubmit()} disabled={loading} style={{
                padding: "10px 24px", backgroundColor: t.accent || "#58a6ff", border: "none",
                borderRadius: 4, color: "#fff", fontFamily: t.font, fontSize: 14, cursor: "pointer", fontWeight: "bold",
                opacity: loading ? 0.5 : 1,
              }}>{loading ? "..." : "OK"}</button>
              <button onClick={() => runInterpreter(findParentMenu(currentProc), {})} style={{
                padding: "10px 24px", backgroundColor: "transparent", border: `1px solid ${t.border || "#30363d"}`,
                borderRadius: 4, color: t.textMuted || "#8b949e", fontFamily: t.font, fontSize: 14, cursor: "pointer",
              }}>Cancel</button>
            </div>
          )}
          {screen?.fields && screen.fields.length > 0 && screen.fields.length === 1 && (screen.lines || []).some(l => /^\s*\d+\.\s/.test(l.text)) && (
            <button onClick={() => handleSubmit()} disabled={loading} style={{
              marginTop: 12, padding: "10px 24px", backgroundColor: t.accent || "#58a6ff", border: "none",
              borderRadius: 4, color: "#fff", fontFamily: t.font, fontSize: 14, cursor: "pointer", fontWeight: "bold",
              opacity: loading ? 0.5 : 1,
            }}>{loading ? "..." : "OK"}</button>
          )}

          {(screen?.done || ((screen?.lines?.length ?? 0) > 0 && (screen?.fields?.length ?? 0) === 0)) && (
            <div style={{ marginTop: 12 }}>
              <button onClick={() => handleSubmit()} style={{
                marginTop: 8, padding: "8px 20px", backgroundColor: t.accent || "#58a6ff", border: "none",
                borderRadius: 4, color: "#fff", fontFamily: t.font, fontSize: 13, cursor: "pointer",
              }}>OK</button>
            </div>
          )}
        </div>
      </main>

      <footer style={{ textAlign: "center", padding: 12, fontSize: 11, color: t.textMuted || "#8b949e", borderTop: `1px solid ${t.border || "#30363d"}` }}>
        db4web — xBase interpreter for the web
      </footer>
    </div>
  );
}

function TableWithScroll({ table, theme: t, onRowAction }: { table: TableData; theme: Record<string, string>; onRowAction: (act: RowAction, row: string[], keyCol: number) => void }) {
  const [sortCol, setSortCol] = React.useState<string>("");
  const [sortDir, setSortDir] = React.useState<string>("asc");
  const [rows, setRows] = React.useState<string[][]>(table.rows);
  const [offset, setOffset] = React.useState(table.offset || 0);
  const [loading, setLoading] = React.useState(false);
  const [hasMore, setHasMore] = React.useState((table.total || 0) > (table.offset || 0) + (table.rows?.length || 0));
  const scrollRef = React.useRef<HTMLDivElement>(null);

  const visibleCols = table.columns.filter(c => c.name.toUpperCase() !== "ID");

  const loadMore = async (newSort?: string, newDir?: string) => {
    if (!table.query || loading) return;
    setLoading(true);
    const s = newSort ?? sortCol;
    const d = newDir ?? sortDir;
    const newOffset = newSort ? 0 : offset + rows.length;
    try {
      const res = await fetch("/api/page", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query: table.query, limit: 50, offset: newOffset, sort: s, dir: d }),
      });
      const data = await res.json();
      const newRows: string[][] = (data.rows || []).map((r: any) => table.columns.map(c => String(r[c.name] ?? "")));
      if (newSort) {
        setRows(newRows);
        setOffset(0);
      } else {
        setRows(prev => [...prev, ...newRows]);
        setOffset(newOffset);
      }
      setHasMore(newRows.length >= 50);
      if (newSort) { setSortCol(newSort); setSortDir(newDir || "asc"); }
    } finally { setLoading(false); }
  };

  const handleSort = (colName: string) => {
    const col = table.columns.find(c => c.name.toUpperCase() !== "ID" && c.name === colName);
    if (!col) return;
    const newDir = sortCol === colName && sortDir === "asc" ? "desc" : "asc";
    loadMore(colName, newDir);
  };

  const handleScroll = () => {
    if (!scrollRef.current || !hasMore || loading) return;
    const { scrollTop, scrollHeight, clientHeight } = scrollRef.current;
    if (scrollHeight - scrollTop - clientHeight < 100) {
      loadMore();
    }
  };

  return (
    <div style={{ marginTop: 8 }}>
      <div ref={scrollRef} onScroll={handleScroll} style={{ maxHeight: 500, overflowY: "auto", overflowX: "hidden" }}>
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
          <thead style={{ position: "sticky", top: 0, zIndex: 1 }}>
            <tr>
              {visibleCols.map((col, i) => (
                <th key={i} onClick={() => handleSort(col.name)} style={{
                  textAlign: col.align === "right" ? "right" : col.align === "center" ? "center" : "left",
                  padding: "8px 12px", borderBottom: `2px solid ${t.border || "#30363d"}`,
                  color: t.accent || "#58a6ff", fontSize: 11, textTransform: "uppercase", letterSpacing: 1,
                  whiteSpace: "nowrap", cursor: "pointer", userSelect: "none",
                  backgroundColor: t.surface || "#161b22",
                }}>
                  {col.name}
                  {sortCol === col.name && <span style={{ marginLeft: 4 }}>{sortDir === "asc" ? "▲" : "▼"}</span>}
                </th>
              ))}
              {(table.actions?.length ?? 0) > 0 && (
                <th style={{ padding: "8px 12px", borderBottom: `2px solid ${t.border || "#30363d"}`, color: t.accent, fontSize: 11, backgroundColor: t.surface || "#161b22" }}>Actions</th>
              )}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, ri) => (
              <tr key={ri}>
                {visibleCols.map((col, ci) => {
                  const dataIdx = table.columns.findIndex(c => c.name === col.name);
                  return (
                    <td key={ci} style={{
                      textAlign: col.align === "right" ? "right" : col.align === "center" ? "center" : "left",
                      padding: "6px 12px", borderBottom: `1px solid ${t.border || "#30363d"}`,
                      color: t.text || "#c9d1d9", fontSize: 13,
                      whiteSpace: "normal", wordBreak: "break-word",
                    }}>{dataIdx >= 0 ? row[dataIdx] : ""}</td>
                  );
                })}
                {(table.actions?.length ?? 0) > 0 && (
                  <td style={{ padding: "6px 12px", borderBottom: `1px solid ${t.border || "#30363d"}`, whiteSpace: "nowrap" }}>
                    {table.actions!.map((act, ai) => (
                      <button key={ai} onClick={() => onRowAction(act, row, table.keyCol ?? 0)} style={{
                        marginRight: 4, padding: "4px 14px", fontSize: 12, borderRadius: 4, cursor: "pointer",
                        border: "none", color: "#fff",
                        backgroundColor: act.label === "Delete" ? (t.accentRed || "#da3633") : (t.accent || "#58a6ff"),
                        fontFamily: t.font, fontWeight: "bold",
                      }}>{act.label}</button>
                    ))}
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
        {loading && <div style={{ padding: 12, textAlign: "center", color: t.textMuted }}>Loading...</div>}
        {!hasMore && rows.length > 0 && <div style={{ padding: 12, textAlign: "center", color: t.textMuted, fontSize: 11 }}>All {table.total || rows.length} records loaded.</div>}
        {rows.length === 0 && <div style={{ padding: 16, textAlign: "center", color: t.textMuted, fontStyle: "italic" }}>No records found.</div>}
      </div>
    </div>
  );
}

const root = document.getElementById("root");
if (root) { ReactDOM.createRoot(root).render(<React.StrictMode><App /></React.StrictMode>); }
