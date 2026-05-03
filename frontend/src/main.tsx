import React from "react";
import ReactDOM from "react-dom/client";

interface ScreenLine { row?: number; col?: number; text: string }
interface ScreenField { row?: number; col?: number; var: string; picture?: string; value?: string; type: string; options?: { value: string; label: string; search?: string }[] }
interface TableColumn { name: string; align?: string }
interface RowAction { label: string; procedure: string }
interface TableData { columns: TableColumn[]; rows: string[][]; actions?: RowAction[]; keyCol?: number; query?: string; limit?: number; offset?: number; total?: number; searchCols?: string[] }
interface Screen { lines: ScreenLine[]; fields: ScreenField[]; prompt?: string; confirm?: string; menu?: MenuDef; wait?: boolean; done?: boolean; result?: string; table?: TableData; title?: string; tagline?: string; nav?: Record<string, string> }
interface MenuItem { label: string; procedure: string }
interface MenuDef { title: string; items: MenuItem[] }

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

function App() {
  const [theme, setTheme] = React.useState<Record<string, string> | null>(null);
  const [screen, setScreen] = React.useState<Screen | null>(null);
  const [fieldVals, setFieldVals] = React.useState<Record<string, string>>({});
  const [loading, setLoading] = React.useState(false);
  const [currentProc, setCurrentProc] = React.useState(MAIN_MENU_PROC);
  const [procStack, setProcStack] = React.useState<string[]>([]);
  const [navStack, setNavStack] = React.useState<string[]>([]);
  const [appTitle, setAppTitle] = React.useState("db4web");
  const [appTagline, setAppTagline] = React.useState("");
  const [programPath, setProgramPath] = React.useState("examples/cureforwoke/app.prg");

  const goBack = () => {
    if (navStack.length > 0) {
      const prev = navStack[navStack.length - 1];
      setNavStack(prev => prev.slice(0, -1));
      setCurrentProc(prev);
    }
  };

  React.useEffect(() => {
    fetch("/api/workspace").then(r => r.json()).then((c: any) => {
      setTheme(c.theme);
      if (c.program) setProgramPath(c.program);
    }).catch(() => setTheme({
      background: "#0d1117", surface: "#161b22", text: "#c9d1d9", textMuted: "#8b949e",
      accent: "#58a6ff", accentGreen: "#3fb950", accentRed: "#da3633", border: "#30363d",
      font: "'Courier New', monospace",
    }));
    runInterpreter(MAIN_MENU_PROC, {});
    // Handle browser back button — navigate within app history
    const handlePop = () => {
      if (procStack.length > 0) {
        const prev = procStack[procStack.length - 1];
        setProcStack(p => p.slice(0, -1));
        setCurrentProc(prev);
        runInterpreter(prev, {});
      }
    };
    window.addEventListener("popstate", handlePop);
    return () => window.removeEventListener("popstate", handlePop);
  }, []);

  const runInterpreter = async (proc: string, input: Record<string, string>) => {
    setLoading(true);
    // Push history state for browser back button support
    if (proc !== currentProc) {
      window.history.pushState({ proc: currentProc }, "", "");
    }
    setCurrentProc(proc);
    setFieldVals({});
    try {
      const res = await fetch("/api/execute", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ program: programPath, procedure: proc, state: { input } }),
      });
      const data = await res.json();
      const s = data.screen as Screen;
      setScreen(s);
      const init: Record<string, string> = {};
      s?.fields?.forEach((f: ScreenField) => { init[f.var] = f.value || ""; });
      // Preserve input values that aren't form fields (like mId, _confirm)
      for (const k of Object.keys(input)) {
        if (init[k] === undefined) init[k] = input[k];
      }
      if (Object.keys(init).length > 0) setFieldVals(init);
      if (s?.title) setAppTitle(s.title);
      if (s?.tagline) setAppTagline(s.tagline);
    } finally { setLoading(false); }
  };

  const [toast, setToast] = React.useState<string>("");

  const handleRowAction = (action: RowAction, row: string[], keyCol: number) => {
    const key = row[keyCol] || "";

    // Delete actions: confirm via overlay, then API call, then refresh list
    if (action.label === "Delete") {
      const table = action.procedure.includes("Service") ? "services"
        : action.procedure.includes("Appt") ? "appointments"
        : action.procedure.includes("Invoice") ? "invoices"
        : "customers";
      setScreen({ ...screen!, confirm: `Delete this ${table.slice(0, -1)}?` });
      // Store the pending delete in a ref/state so the confirm handler knows what to do
      setPendingDelete({ table, key, action, rowKey: key });
      return;
    }

    // Edit actions: build input from row data
    const input: Record<string, string> = { mId: key };
    const tbl = action.procedure.includes("Service") ? "services" : "customers";
    if (tbl === "customers" && row.length >= 5) {
      input.mName = row[1] || "";
      input.mAlias = row[2] || "";
      input.mPhone = row[3] || "";
      input.mEmail = "";
    }
    if (tbl === "services" && row.length >= 5) {
      input.mName = row[1] || "";
      // Don't pass other fields — they're used as save triggers in EditService
    }
    if (tbl === "services" && row.length >= 4) {
      input.mName = row[1] || "";
      input.mDesc = row[2] || "";
      input.mPrice = row[3] || "";
    }
    runInterpreter(action.procedure, input);
  };

  const [pendingDelete, setPendingDelete] = React.useState<{ table: string; key: string; action: RowAction; rowKey: string } | null>(null);

  const handleSubmit = async (directChoice?: string, confirmResult?: string) => {
    if (!screen) return;

    // Handle confirmation response for pending deletes
    if (screen.confirm && pendingDelete) {
      if (confirmResult === "yes") {
        try {
          await fetch(`/api/data/${pendingDelete.table}/${pendingDelete.key}`, { method: "DELETE" });
          setToast("Deleted!");
          setTimeout(() => setToast(""), 2000);
          setPendingDelete(null);
          const savedProc = currentProc;
          const res = await fetch("/api/execute", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ program: programPath, procedure: savedProc, state: {} }),
          });
          const data = await res.json();
          const s = data.screen as Screen;
          setScreen(s);
          const init: Record<string, string> = {};
          s?.fields?.forEach((f: ScreenField) => { init[f.var] = f.value || ""; });
          if (Object.keys(init).length > 0) setFieldVals(init);
        } catch { setToast("Delete failed"); }
      } else {
        setPendingDelete(null);
        setScreen({ ...screen!, confirm: "" });
      }
      return;
    }

    // Handle confirmation response from .prg CONFIRM
    if (screen.confirm) {
      runInterpreter(currentProc, { ...fieldVals, _confirm: confirmResult || "yes" });
      return;
    }

    // Menu item click — navigate directly
    if (screen.menu && directChoice) {
      // For AddAppointment, handle in frontend (fetch options, show autocomplete)
      if (directChoice === "AddAppointment") {
        setProcStack(prev => [...prev, currentProc]);
        setCurrentProc("AddAppointment");
        setFieldVals({ mCustomer: "", mService: "", mScheduled: "" });
        try {
          const [custData, svcData] = await Promise.all([
            fetch("/api/data/customers").then(r => r.json()),
            fetch("/api/data/services").then(r => r.json()),
          ]);
          const custOpts = (custData.rows || []).map((c: any) => ({ value: String(c.id), label: `${c.name || ""}${c.alias ? " (" + c.alias + ")" : ""}`, search: `${c.name || ""} ${c.alias || ""} ${c.phone || ""}` }));
          const svcOpts = (svcData.rows || []).map((s: any) => ({ value: String(s.id), label: `${s.name || ""} — $${parseFloat(s.base_price || "0").toFixed(0)}`, search: `${s.name || ""} ${s.description || ""}` }));
          setScreen({
            lines: [{ row: 1, col: 1, text: "Schedule Appointment" }],
            fields: [
              { var: "mCustomer", type: "get", picture: "9999", options: custOpts },
              { var: "mService", type: "get", picture: "99", options: svcOpts },
              { var: "mScheduled", type: "get", picture: "9999-99-99 99:99" },
            ],
            prompt: "mCustomer",
          });
        } catch { setScreen({ lines: [{ row: 1, col: 1, text: "Failed to load" }], fields: [], done: true }); }
        return;
      }
      setProcStack(prev => [...prev, currentProc]);
      runInterpreter(directChoice, {});
      return;
    }

    // Wait/done screens: return to parent menu
    if (screen.done || screen.wait) {
      if (procStack.length > 0) {
        const parent = procStack[procStack.length - 1];
        setProcStack(prev => prev.slice(0, -1));
        runInterpreter(parent, {});
      } else {
        runInterpreter(MAIN_MENU_PROC, {});
      }
      return;
    }

    if (screen.done) {
      // Procedure finished — go back to caller via procStack
      if (procStack.length > 0) {
        const parent = procStack[procStack.length - 1];
        setProcStack(prev => prev.slice(0, -1));
        runInterpreter(parent, {});
      } else {
        runInterpreter(MAIN_MENU_PROC, {});
      }
      return;
    }
    const choice = directChoice ?? (screen.prompt ? (fieldVals[screen.prompt] || "") : "");

    // Navigation from NAV statement (defined in .prg)
    if (screen.nav && screen.nav[choice]) {
      const target = screen.nav[choice];
      if (target === "QUIT" || target === "MainMenu") {
        setScreen({ lines: [{ row: 1, col: 1, text: target === "QUIT" ? "Goodbye!" : "" }], fields: [], done: true });
        return;
      }
      // Push current menu onto procStack so back/return works
      setProcStack(prev => [...prev, currentProc]);
      runInterpreter(target, {});
      return;
    }

    // AddAppointment save via API
    if (currentProc === "AddAppointment" && fieldVals.mCustomer) {
      setLoading(true);
      try {
        const svc = await(await fetch(`/api/data/services/${fieldVals.mService || "1"}`)).json();
        const duration = parseInt(svc.duration) || 60;
        await fetch("/api/data/appointments", { method: "POST", headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ customer_id: parseInt(fieldVals.mCustomer), service_id: parseInt(fieldVals.mService) || 1, scheduled_for: fieldVals.mScheduled || new Date().toISOString(), duration, status: "pending" }) });
        setScreen({ lines: [{ row: 1, col: 1, text: "Appointment scheduled!" }], fields: [], done: true });
        setTimeout(() => {
          const parent = procStack.length > 0 ? procStack[procStack.length - 1] : MAIN_MENU_PROC;
          runInterpreter(parent, {});
        }, 600);
      } catch (e: any) { setScreen({ lines: [{ row: 1, col: 1, text: `Error: ${e.message}` }], fields: [], done: true }); }
      finally { setLoading(false); }
      return;
    }

    // No nav mapping — mark as form submission and send back
    runInterpreter(currentProc, { ...fieldVals, _save: "yes" });
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

      {procStack.length > 0 && (
        <div style={{ backgroundColor: t.surface || "#161b22", borderBottom: `1px solid ${t.border || "#30363d"}`, padding: "6px 24px", display: "flex", alignItems: "center", gap: 8, fontSize: 12, fontFamily: t.font }}>
          <button onClick={goBack} style={{
            background: "none", border: `1px solid ${t.border || "#30363d"}`, borderRadius: 4, color: t.accent || "#58a6ff",
            fontFamily: t.font, fontSize: 12, cursor: "pointer", padding: "2px 10px",
          }}>&larr;</button>
          <span style={{ color: t.textMuted || "#8b949e" }}>
            {[...procStack, currentProc].map((v, i) => {
              const isLast = i === procStack.length;
              return (
              <span key={i}>
                {i > 0 && <span style={{ margin: "0 4px", color: t.textMuted }}>/</span>}
                <span
                  onClick={() => {
                    if (isLast) return;
                    const target = v;
                    setProcStack(prev => prev.slice(0, i));
                    setCurrentProc(target);
                    runInterpreter(target, {});
                  }}
                  style={{
                    color: isLast ? t.text : t.accent,
                    cursor: isLast ? "default" : "pointer",
                    textDecoration: isLast ? "none" : "underline",
                  }}
                >{v.replace(/([A-Z])/g, ' $1').trim()}</span>
              </span>
            )})}
          </span>
        </div>
      )}

      <main style={{ flex: 1, padding: 24, maxWidth: "95%", width: "100%", margin: "0 auto", boxSizing: "border-box" }}>
        <div style={{ marginBottom: 16 }}>
          <h2 style={{ margin: 0, fontSize: 18, color: t.text || "#c9d1d9", fontWeight: "bold" }}>
            {currentProc.replace(/([A-Z])/g, ' $1').replace(/^(.)/, c => c.toUpperCase()).trim()}
          </h2>
          {loading && <span style={{ fontSize: 11, color: t.accent }}> executing...</span>}
        </div>

        <div style={{ backgroundColor: "#0a0a0f", border: `1px solid ${t.border || "#30363d"}`, borderRadius: 8, padding: 24, fontSize: 14, lineHeight: 1.6 }}>
          {/* Declarative menu rendering */}
          {screen?.menu && (
            <div>
              {screen.menu.items.map((item, i) => (
                <div key={i} onClick={() => handleSubmit(item.procedure)}
                  style={{
                    padding: "8px 12px", margin: "2px 0", borderRadius: 4, cursor: "pointer",
                    color: t.accent || "#58a6ff", fontSize: 14,
                    transition: "background 0.1s",
                  }}
                  onMouseEnter={e => e.currentTarget.style.background = t.surface || "#161b22"}
                  onMouseLeave={e => e.currentTarget.style.background = "transparent"}
                >{item.label}</div>
              ))}
            </div>
          )}

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
            <TableWithScroll table={screen.table} theme={t} onRowAction={handleRowAction} highlightKey={pendingDelete?.rowKey} />
          )}

          {(screen?.fields || []).filter(() => {
            return !(screen?.lines?.some?.(l => /^\s*\d+\.\s/.test(l.text)) && screen.fields?.length === 1);
          }).map((f, i) => (
            <div key={i} style={{ marginTop: 8 }}>
              <div style={{ fontSize: 11, color: t.textMuted, marginBottom: 4 }}>{f.var.replace(/^m/, "")}</div>
              {f.options ? (
                <AutoComplete
                  options={f.options}
                  value={fieldVals[f.var] ?? ""}
                  onChange={(val) => setFieldVals({ ...fieldVals, [f.var]: val })}
                  theme={t}
                  autoFocus={i === 0}
                />
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
              <button onClick={() => {
                const parent = procStack.length > 0 ? procStack[procStack.length - 1] : MAIN_MENU_PROC;
                setProcStack(prev => prev.slice(0, -1));
                runInterpreter(parent, {});
              }} style={{
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

      {screen?.confirm && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, backgroundColor: pendingDelete ? (t.surface || "#161b22") : (t.surface || "#161b22"), borderBottom: `2px solid ${pendingDelete ? (t.accentRed || "#da3633") : (t.accent || "#58a6ff")}`, padding: "16px 24px", zIndex: 1000, display: "flex", alignItems: "center", justifyContent: "center", gap: 16 }}>
          <span style={{ color: t.text, fontSize: 14 }}>{screen.confirm}</span>
          <button onClick={() => handleSubmit(undefined, "yes")} style={{ padding: "6px 20px", backgroundColor: pendingDelete ? (t.accentRed || "#da3633") : (t.accent || "#58a6ff"), border: "none", borderRadius: 4, color: "#fff", cursor: "pointer", fontFamily: t.font, fontSize: 13, fontWeight: "bold" }}>Yes</button>
          <button onClick={() => handleSubmit(undefined, "no")} style={{ padding: "6px 20px", backgroundColor: "transparent", border: `1px solid ${t.border}`, borderRadius: 4, color: t.textMuted, cursor: "pointer", fontFamily: t.font, fontSize: 13 }}>No</button>
        </div>
      )}

      {toast && (
        <div style={{ position: "fixed", top: 60, left: "50%", transform: "translateX(-50%)", backgroundColor: t.accentGreen || "#3fb950", padding: "10px 24px", borderRadius: 8, zIndex: 1001, color: "#fff", fontSize: 14, fontWeight: "bold", fontFamily: t.font }}>
          {toast}
        </div>
      )}

      <footer style={{ textAlign: "center", padding: 12, fontSize: 11, color: t.textMuted || "#8b949e", borderTop: `1px solid ${t.border || "#30363d"}` }}>
        db4web — xBase interpreter for the web
      </footer>
    </div>
  );
}

function TableWithScroll({ table, theme: t, onRowAction, highlightKey }: { table: TableData; theme: Record<string, string>; onRowAction: (act: RowAction, row: string[], keyCol: number) => void; highlightKey?: string }) {
  const [rows, setRows] = React.useState<string[][]>(table.rows);
  const [offset, setOffset] = React.useState(table.offset || 0);
  const [sortCol, setSortCol] = React.useState<string>("");
  const [sortDir, setSortDir] = React.useState<string>("asc");
  const [loading, setLoading] = React.useState(false);
  const [hasMore, setHasMore] = React.useState((table.total || 0) > (table.offset || 0) + (table.rows?.length || 0));
  const [searchText, setSearchText] = React.useState<string>("");
  const [searchCols, setSearchCols] = React.useState<Set<string>>(new Set(table.searchCols?.slice(0, 1) || []));
  const [showSearchDropdown, setShowSearchDropdown] = React.useState(false);
  const scrollRef = React.useRef<HTMLDivElement>(null);
  const searchTimer = React.useRef<ReturnType<typeof setTimeout>>();
  const visibleCols = table.columns.filter(c => c.name.toUpperCase() !== "ID");

  // Sync rows when table data changes (e.g. after delete)
  React.useEffect(() => {
    setRows(table.rows);
    setOffset(table.offset || 0);
    setSortCol("");
    setSortDir("asc");
    setHasMore((table.total || 0) > (table.offset || 0) + (table.rows?.length || 0));
    if ((table.searchCols || []).length > 0 && searchCols.size === 0) {
      setSearchCols(new Set([(table.searchCols || [])[0]]));
    }
  }, [table]);

  // Close search dropdown on outside click
  React.useEffect(() => {
    if (!showSearchDropdown) return;
    const handler = () => setShowSearchDropdown(false);
    window.addEventListener("click", handler);
    return () => window.removeEventListener("click", handler);
  }, [showSearchDropdown]);

  const doSearch = (text: string) => {
    if (!table.query) return;
    setLoading(true);
    const sql = table.query;
    const cols = Array.from(searchCols);
    fetch("/api/page", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ query: sql, limit: 50, offset: 0, sort: sortCol, dir: sortDir, search: text, searchCol: cols.join(",") }),
    }).then(r => r.json()).then(data => {
      const newRows: string[][] = (data.rows || []).map((r: any) => table.columns.map(c => String(r[c.name] ?? r[c.name.toLowerCase()] ?? r[c.name.toUpperCase()] ?? "")));
      setRows(newRows);
      setOffset(0);
      setHasMore(newRows.length >= 50);
      setLoading(false);
    });
  };

  const handleSearchChange = (text: string) => {
    setSearchText(text);
    clearTimeout(searchTimer.current);
    searchTimer.current = setTimeout(() => doSearch(text), 300);
  };

  const toggleSearchCol = (col: string) => {
    const next = new Set(searchCols);
    if (next.has(col)) next.delete(col); else next.add(col);
    if (next.size === 0) next.add(col); // keep at least one
    setSearchCols(next);
    clearTimeout(searchTimer.current);
    searchTimer.current = setTimeout(() => doSearch(searchText), 300);
  };

  const clearSearch = () => {
    setSearchText("");
    doSearch("");
  };

  const loadMore = async (newSort?: string, newDir?: string) => {
    if (!table.query || loading) return;
    setLoading(true);
    const s = newSort ?? sortCol;
    const d = newDir ?? sortDir;
    const nextOffset = newSort ? 0 : offset + 50;
    try {
      const res = await fetch("/api/page", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query: table.query, limit: 50, offset: nextOffset, sort: s, dir: d, search: searchText, searchCol: Array.from(searchCols).join(",") }),
      });
      const data = await res.json();
      const newRows: string[][] = (data.rows || []).map((r: any) => table.columns.map(c => {
        // API returns lowercase SQL column names, but COLUMNS in .prg may be uppercase
        const val = r[c.name] ?? r[c.name.toLowerCase()] ?? r[c.name.toUpperCase()] ?? "";
        return String(val);
      }));
      if (newSort) {
        setRows(newRows);
        setOffset(0);
      } else {
        setRows(prev => [...prev, ...newRows]);
        setOffset(nextOffset);
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
      {/* Search bar */}
      {(table.searchCols?.length ?? 0) > 0 && (
        <div style={{ display: "flex", gap: 8, marginBottom: 8, alignItems: "center" }}>
          <div style={{ position: "relative" }}>
            <button onClick={e => { e.stopPropagation(); setShowSearchDropdown(!showSearchDropdown); }}
              style={{
                backgroundColor: t.background, border: `1px solid ${t.border}`, borderRadius: 4,
                padding: "6px 10px", fontFamily: t.font, fontSize: 13, color: t.text, cursor: "pointer",
                whiteSpace: "nowrap",
              }}>
              Search: {Array.from(searchCols).join(", ") || "none"} ▾
            </button>
            {showSearchDropdown && (
              <div style={{ position: "absolute", top: "100%", left: 0, backgroundColor: t.surface || "#161b22", border: `1px solid ${t.border}`, borderRadius: 4, zIndex: 10, minWidth: 150, marginTop: 2 }}>
                {(table.searchCols ?? []).map((sc) => (
                  <label key={sc} onClick={e => e.stopPropagation()} style={{ display: "flex", alignItems: "center", gap: 6, padding: "6px 10px", cursor: "pointer", fontSize: 13, color: t.text }}
                    onMouseEnter={e => e.currentTarget.style.background = t.background || "#0d1117"}
                    onMouseLeave={e => e.currentTarget.style.background = "transparent"}
                  >
                    <input type="checkbox" checked={searchCols.has(sc)} onChange={() => toggleSearchCol(sc)} />
                    {sc}
                  </label>
                ))}
              </div>
            )}
          </div>
          <div style={{ position: "relative", flex: 1, maxWidth: 300 }}>
            <input value={searchText} onChange={e => handleSearchChange(e.target.value)}
              placeholder="Search..." style={{
                width: "100%", boxSizing: "border-box",
                backgroundColor: t.background, border: `1px solid ${t.border}`, borderRadius: 4,
                padding: "6px 30px 6px 10px", fontFamily: t.font, fontSize: 13, color: t.text, outline: "none",
              }} />
            {searchText && (
              <span onClick={clearSearch} style={{ position: "absolute", right: 8, top: "50%", transform: "translateY(-50%)", cursor: "pointer", color: t.textMuted, fontSize: 16 }}>&times;</span>
            )}
          </div>
        </div>
      )}
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
            {rows.map((row, ri) => {
              const rowId = row[table.keyCol ?? 0];
              const isHighlighted = highlightKey && rowId === highlightKey;
              return (
              <tr key={ri} style={{ backgroundColor: isHighlighted ? (t.accentRed ? t.accentRed + "22" : "#330000") : undefined }}>
                {visibleCols.map((col, ci) => {
                  const dataIdx = table.columns.findIndex(c => c.name === col.name);
                  return (
                    <td key={ci} style={{
                      textAlign: col.align === "right" ? "right" : col.align === "center" ? "center" : "left",
                      padding: "6px 12px", borderBottom: `1px solid ${t.border || "#30363d"}`,
                      color: t.text || "#c9d1d9", fontSize: 13,
                      whiteSpace: col.name === "Description" ? "normal" : "nowrap",
                      wordBreak: col.name === "Description" ? "break-word" : undefined,
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
            );
          })}
          </tbody>
        </table>
        {loading && <div style={{ padding: 12, textAlign: "center", color: t.textMuted }}>Loading...</div>}
        {!hasMore && rows.length > 0 && <div style={{ padding: 12, textAlign: "center", color: t.textMuted, fontSize: 11 }}>All {table.total || rows.length} records loaded.</div>}
        {rows.length === 0 && <div style={{ padding: 16, textAlign: "center", color: t.textMuted, fontStyle: "italic" }}>No records found.</div>}
      </div>
    </div>
  );
}

function AutoComplete({ options, value, onChange, theme: t, autoFocus }: {
  options: { value: string; label: string; search?: string }[];
  value: string;
  onChange: (v: string) => void;
  theme: Record<string, string>;
  autoFocus?: boolean;
}) {
  const [text, setText] = React.useState(options.find(o => o.value === value)?.label ?? "");
  const [open, setOpen] = React.useState(false);
  const [focusIdx, setFocusIdx] = React.useState(-1);
  const ref = React.useRef<HTMLDivElement>(null);
  const inputRef = React.useRef<HTMLInputElement>(null);
  const listRef = React.useRef<HTMLDivElement>(null);

  const sorted = [...options].sort((a, b) => a.label.localeCompare(b.label));
  const filtered = text
    ? sorted.filter(o => {
        const q = text.toLowerCase();
        const searchText = (o.search || o.label).toLowerCase();
        return searchText.includes(q);
      })
    : sorted;

  React.useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => { if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false); };
    window.addEventListener("mousedown", handler);
    return () => window.removeEventListener("mousedown", handler);
  }, [open]);

  React.useEffect(() => {
    const cur = options.find(o => o.value === value);
    setText(cur?.label ?? "");
  }, [value, options]);

  const select = (idx: number) => {
    if (idx < 0 || idx >= filtered.length) return;
    onChange(filtered[idx].value);
    setText(filtered[idx].label);
    setOpen(false);
    setFocusIdx(-1);
  };

  const keyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      setOpen(false);
      setFocusIdx(-1);
      inputRef.current?.blur();
      return;
    }
    if ((e.ctrlKey || e.metaKey) && e.key === "ArrowDown") {
      e.preventDefault();
      setOpen(true);
      setFocusIdx(-1);
      return;
    }
    if (!open) {
      if (e.key === "Tab") return; // normal tab behavior
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setOpen(true);
        return;
      }
      return;
    }
    // Dropdown is open
    if (e.key === "Tab") {
      if (e.shiftKey) {
        // Shift+Tab: close and focus previous
        setOpen(false);
        return;
      }
      e.preventDefault();
      if (document.activeElement === inputRef.current) {
        // Focus dropdown list
        listRef.current?.focus();
        return;
      }
      // Focused on dropdown → close and tab to next
      setOpen(false);
      // Synthesize a tab keypress on the input
      inputRef.current?.focus();
      return;
    }
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setFocusIdx(i => Math.min(i + 1, filtered.length - 1));
      return;
    }
    if (e.key === "ArrowUp") {
      e.preventDefault();
      setFocusIdx(i => Math.max(i - 1, 0));
      return;
    }
    if (e.key === "Enter") {
      e.preventDefault();
      if (focusIdx >= 0) { select(focusIdx); return; }
      if (filtered.length === 1) { select(0); return; }
    }
  };

  return (
    <div ref={ref} style={{ position: "relative" }}>
      <div style={{ display: "flex" }}>
        <input ref={inputRef} autoFocus={autoFocus} value={text}
          onChange={e => { setText(e.target.value); setOpen(true); setFocusIdx(-1); }}
          onFocus={() => {/* don't auto-open */}}
          onKeyDown={keyDown}
          placeholder="Search..."
          style={{
            flex: 1, backgroundColor: t.background, border: `1px solid ${t.border}`, borderRadius: "4px 0 0 4px",
            padding: "10px 14px", fontFamily: t.font, fontSize: 16, color: t.text, outline: "none",
          }} />
        <button onClick={() => { setOpen(!open); if (!open) inputRef.current?.focus(); }}
          style={{
            backgroundColor: t.background, border: `1px solid ${t.border}`, borderLeft: "none", borderRadius: "0 4px 4px 0",
            padding: "0 12px", cursor: "pointer", color: t.textMuted, fontSize: 12, fontFamily: t.font,
          }}>▼</button>
      </div>
      {open && (
        <div ref={listRef} tabIndex={-1} onKeyDown={keyDown}
          style={{ position: "absolute", top: "100%", left: 0, right: 0, backgroundColor: t.surface || "#161b22", border: `1px solid ${t.border}`, borderRadius: 4, zIndex: 20, maxHeight: 300, overflowY: "auto", marginTop: 2, outline: "none" }}>
          {filtered.map((o, i) => (
            <div key={i} onClick={() => select(i)}
              onMouseEnter={() => setFocusIdx(i)}
              style={{
                padding: "8px 12px", cursor: "pointer", fontSize: 14, color: t.text,
                backgroundColor: i === focusIdx ? (t.accent + "22") : "transparent",
                borderBottom: i < filtered.length - 1 ? `1px solid ${t.border}` : "none",
              }}>{o.label}</div>
          ))}
          {filtered.length === 0 && <div style={{ padding: "8px 12px", color: t.textMuted, fontSize: 13, fontStyle: "italic" }}>No matches</div>}
        </div>
      )}
    </div>
  );
}

const root = document.getElementById("root");
if (root) { ReactDOM.createRoot(root).render(<React.StrictMode><App /></React.StrictMode>); }
