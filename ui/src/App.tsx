import { useState, useEffect } from 'react';
import { Server, Play, Square, Trash2, ArrowRightLeft, Shield, Loader2, Search, Edit2, Box } from 'lucide-react';
import { DockerAPI } from './services/api';
import { useMigrationWebSocket } from './hooks/useMigrationWebSocket';
import type { HostConfig, DockerContainer } from './types/docker';

export default function App() {
  // --- STATE ---
  const [host, setHost] = useState<HostConfig>(() => JSON.parse(localStorage.getItem('mainHost') || '{"ip":"localhost:8080","token":""}'));
  const [containers, setContainers] = useState<DockerContainer[]>([]);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(false);

  // Modals State
  const [migrateModal, setMigrateModal] = useState<{ isOpen: boolean, targetIp: string, targetToken: string }>({ isOpen: false, targetIp: '', targetToken: '' });
  const [renameModal, setRenameModal] = useState<{ isOpen: boolean, id: string, currentName: string, newName: string }>({ isOpen: false, id: '', currentName: '', newName: '' });

  const events = useMigrationWebSocket(host);

  // --- EFFECTS ---
  // auto token encapsulation to app via magic link if clicked on terminal
  useEffect(() => {
    // 1. Check if there is a token in the URL (?token=abc...)
    const params = new URLSearchParams(window.location.search);
    const urlToken = params.get('token');

    if (urlToken) {
      // 2. Update the host state with the current window location and the new token
      const newHost = {
        ip: window.location.host, // Automatically gets "localhost:8080" or "192.168...:8080"
        token: urlToken,
        nickname: 'Local Node'
      };

      setHost(newHost);
      localStorage.setItem('mainHost', JSON.stringify(newHost));

      // 3. CLEAN UP: Remove the token from the URL bar for security
      window.history.replaceState({}, document.title, window.location.pathname);

      console.log("🔑 Magic Link detected. Token applied successfully.");
    }
  }, []);

  useEffect(() => {
    localStorage.setItem('mainHost', JSON.stringify(host));
    if (host.token) fetchContainers();
  }, [host.ip, host.token]);

  // --- ACTIONS ---
  const fetchContainers = async () => {
    try {
      setLoading(true);
      setContainers(await DockerAPI.getContainers(host));
    } catch (e) {
      console.error("Failed to fetch");
    } finally {
      setLoading(false);
    }
  };

  const handleAction = async (action: 'start' | 'stop' | 'delete') => {
    if (selectedIds.size === 0) return;
    const force = action === 'delete' && confirm('Force delete running containers?');
    try {
      setLoading(true);
      await DockerAPI.containerAction(host, action, Array.from(selectedIds), force);
      setSelectedIds(new Set()); // Clear selection
      await fetchContainers();
    } catch (e) {
      alert(`Action ${action} failed`);
    } finally {
      setLoading(false);
    }
  };

  const handleRename = async () => {
    try {
      setLoading(true);
      await DockerAPI.renameContainer(host, renameModal.id, renameModal.newName);
      setRenameModal({ ...renameModal, isOpen: false });
      await fetchContainers();
    } catch (e) {
      alert("Rename failed");
    } finally {
      setLoading(false);
    }
  };

  const handleMigrate = async () => {
    try {
      await DockerAPI.migrateBatch(host, migrateModal.targetIp, migrateModal.targetToken, Array.from(selectedIds));
      setMigrateModal({ ...migrateModal, isOpen: false });
      setSelectedIds(new Set()); // Deselect so they can watch the progress bars
    } catch (e) {
      alert("Migration failed to start");
    }
  };

  // --- HELPERS ---
  const toggleSelect = (id: string) => {
    const next = new Set(selectedIds);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelectedIds(next);
  };

  const toggleAll = () => {
    if (selectedIds.size === filteredContainers.length) setSelectedIds(new Set());
    else setSelectedIds(new Set(filteredContainers.map(c => c.Id)));
  };

  const filteredContainers = containers.filter(c => c.Names[0].toLowerCase().includes(search.toLowerCase()));

  // --- RENDER ---
  return (
    <div className="min-h-screen bg-[#020408] text-slate-300 font-sans selection:bg-brand-500/30 pb-20">

      {/* Top Navigation */}
      <nav className="border-b border-slate-800 bg-slate-900/50 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-6 h-16 flex justify-between items-center">
          <div className="flex items-center gap-3">
            {/* The DP Logo Block with a subtle neon glow */}
            <div className="w-8 h-8 bg-brand-500 rounded-lg flex items-center justify-center text-white font-black shadow-[0_0_15px_#0ea5e950] tracking-tighter">
              DP
            </div>

            {/* The Name & Suffix Layout */}
            <div className="flex items-baseline gap-2">
              <h1 className="text-xl font-black tracking-tighter text-white uppercase">
                Dock<span className="text-brand-500">Porter</span>
              </h1>

              {/* The sleek 'by DevLouix' developer badge */}
              <span className="text-[9px] font-bold text-slate-400 uppercase tracking-widest bg-slate-800/50 px-2 py-0.5 rounded-full border border-slate-700/50 relative -top-1">
                by DevLouix
              </span>
            </div>
          </div>

          {/* Host Config Panel inline */}
          <div className="flex items-center gap-3">
            <div className="flex items-center bg-black/40 border border-slate-800 rounded-lg px-2 overflow-hidden focus-within:border-brand-500 transition-colors">
              <Server size={14} className="text-slate-500 mx-2" />
              <input
                className="bg-transparent text-xs py-2 w-32 outline-none text-white placeholder-slate-600"
                placeholder="Agent IP" value={host.ip} onChange={e => setHost({ ...host, ip: e.target.value })}
              />
              <div className="w-px h-4 bg-slate-800 mx-2" />
              <input
                type="password" className="bg-transparent text-xs py-2 w-32 outline-none text-brand-500 placeholder-slate-600"
                placeholder="Auth Token" value={host.token} onChange={e => setHost({ ...host, token: e.target.value })}
              />
            </div>
            <button onClick={fetchContainers} className="p-2 bg-slate-800 hover:bg-brand-500 hover:text-white rounded-lg transition-colors">
              <Loader2 size={16} className={loading ? "animate-spin" : ""} />
            </button>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto p-6 mt-4">

        {/* Toolbar */}
        <div className="flex justify-between items-center mb-6">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={16} />
            <input
              className="bg-slate-900 border border-slate-800 rounded-xl pl-10 pr-4 py-2 text-sm text-white focus:border-brand-500 outline-none w-64 transition-all"
              placeholder="Search containers..." value={search} onChange={e => setSearch(e.target.value)}
            />
          </div>

          <div className="text-xs font-bold text-slate-500 uppercase tracking-widest bg-slate-800/30 px-4 py-2 rounded-lg border border-slate-800/50 flex items-center gap-2">
            <Shield size={14} className="text-emerald-500" /> API Secured
          </div>
        </div>

        {/* Data Table */}
        <div className="bg-slate-900 border border-slate-800 rounded-2xl overflow-hidden shadow-2xl">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-slate-800/50 text-xs uppercase tracking-widest text-slate-500 border-b border-slate-800">
                <th className="p-4 w-12 text-center">
                  <input type="checkbox" className="accent-brand-500 w-4 h-4 rounded cursor-pointer"
                    checked={selectedIds.size > 0 && selectedIds.size === filteredContainers.length}
                    onChange={toggleAll}
                  />
                </th>
                <th className="p-4">Name</th>
                <th className="p-4">Image</th>
                <th className="p-4">Status</th>
                <th className="p-4 text-right">Quick Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/50">
              {filteredContainers.map(c => {
                const name = c.Names[0].replace('/', '');
                const isSelected = selectedIds.has(c.Id);
                const isRunning = c.State === 'running';
                const event = events[c.Id];

                return (
                  <tr key={c.Id} className={`hover:bg-slate-800/20 transition-colors ${isSelected ? 'bg-brand-500/5' : ''}`}>
                    <td className="p-4 text-center">
                      <input type="checkbox" className="accent-brand-500 w-4 h-4 rounded cursor-pointer"
                        checked={isSelected} onChange={() => toggleSelect(c.Id)} disabled={!!event}
                      />
                    </td>
                    <td className="p-4">
                      <div className="flex items-center gap-3">
                        <Box size={16} className={isRunning ? 'text-emerald-500' : 'text-slate-600'} />
                        <span className="font-bold text-white text-sm">{name}</span>
                      </div>
                      {/* Active Migration Progress Bar (If migrating) */}
                      {event && (
                        <div className="mt-2 flex items-center gap-3">
                          <div className="h-1.5 w-full bg-slate-800 rounded-full overflow-hidden">
                            <div className="h-full bg-brand-500 transition-all duration-500" style={{ width: `${event.progress}%` }} />
                          </div>
                          <span className="text-[10px] font-bold uppercase text-brand-500">{event.status}</span>
                        </div>
                      )}
                    </td>
                    <td className="p-4 text-xs font-mono text-slate-400">{c.Image}</td>
                    <td className="p-4">
                      <span className={`px-2.5 py-1 rounded-md text-[10px] font-bold uppercase tracking-widest border ${isRunning ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' : 'bg-slate-800 text-slate-500 border-slate-700'
                        }`}>
                        {c.State}
                      </span>
                    </td>
                    <td className="p-4 text-right">
                      <button onClick={() => setRenameModal({ isOpen: true, id: c.Id, currentName: name, newName: name })} className="p-1.5 text-slate-500 hover:text-white transition-colors" title="Rename">
                        <Edit2 size={16} />
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </main>

      {/* --- MULTI-SELECT FLOATING ACTION BAR --- */}
      {selectedIds.size > 0 && (
        <div className="fixed bottom-8 left-1/2 -translate-x-1/2 bg-slate-900 border border-slate-700 shadow-2xl rounded-2xl px-6 py-4 flex items-center gap-6 z-50 animate-in slide-in-from-bottom-10 fade-in">
          <div className="text-sm font-bold text-white">
            <span className="text-brand-500">{selectedIds.size}</span> Selected
          </div>
          <div className="w-px h-6 bg-slate-700" />
          <div className="flex gap-2">
            <button onClick={() => handleAction('start')} className="flex items-center gap-2 px-3 py-1.5 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-500 rounded-lg text-xs font-bold uppercase tracking-widest transition-colors"><Play size={14} /> Start</button>
            <button onClick={() => handleAction('stop')} className="flex items-center gap-2 px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg text-xs font-bold uppercase tracking-widest transition-colors"><Square size={14} /> Stop</button>
            <button onClick={() => handleAction('delete')} className="flex items-center gap-2 px-3 py-1.5 bg-red-500/10 hover:bg-red-500/20 text-red-500 rounded-lg text-xs font-bold uppercase tracking-widest transition-colors"><Trash2 size={14} /> Delete</button>
          </div>
          <div className="w-px h-6 bg-slate-700" />
          <button onClick={() => setMigrateModal({ ...migrateModal, isOpen: true })} className="flex items-center gap-2 px-4 py-2 bg-brand-600 hover:bg-brand-500 text-white shadow-[0_0_15px_#0ea5e950] rounded-lg text-xs font-bold uppercase tracking-widest transition-all">
            <ArrowRightLeft size={16} /> Migrate
          </button>
        </div>
      )}

      {/* --- RENAME MODAL --- */}
      {renameModal.isOpen && (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-slate-900 border border-slate-800 p-6 rounded-2xl w-96 shadow-2xl">
            <h3 className="text-lg font-bold text-white mb-4">Rename Container</h3>
            <input
              autoFocus
              className="w-full bg-black/50 border border-slate-700 rounded-lg p-3 text-white focus:border-brand-500 outline-none mb-6"
              value={renameModal.newName}
              onChange={e => setRenameModal({ ...renameModal, newName: e.target.value })}
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setRenameModal({ ...renameModal, isOpen: false })} className="px-4 py-2 text-slate-400 hover:text-white font-bold text-sm">Cancel</button>
              <button onClick={handleRename} className="px-4 py-2 bg-brand-600 hover:bg-brand-500 text-white rounded-lg font-bold text-sm shadow-lg">Save Changes</button>
            </div>
          </div>
        </div>
      )}

      {/* --- BATCH MIGRATE MODAL --- */}
      {migrateModal.isOpen && (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-slate-900 border border-brand-500/30 p-8 rounded-3xl w-[450px] shadow-[0_0_50px_#0ea5e920]">
            <div className="flex items-center gap-3 mb-6">
              <div className="p-3 bg-brand-500/10 text-brand-500 rounded-xl"><ArrowRightLeft size={24} /></div>
              <div>
                <h3 className="text-xl font-bold text-white">Ship to Target</h3>
                <p className="text-xs text-slate-500">Migrating {selectedIds.size} containers</p>
              </div>
            </div>

            <div className="space-y-4 mb-8">
              <div>
                <label className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1 block">Target IP Address</label>
                <input
                  className="w-full bg-black/50 border border-slate-700 rounded-xl p-3 text-sm font-mono text-white focus:border-brand-500 outline-none"
                  placeholder="e.g. 192.168.1.50:8080"
                  value={migrateModal.targetIp} onChange={e => setMigrateModal({ ...migrateModal, targetIp: e.target.value })}
                />
              </div>
              <div>
                <label className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1 block">Target Auth Token</label>
                <input
                  type="password"
                  className="w-full bg-black/50 border border-slate-700 rounded-xl p-3 text-sm font-mono text-brand-500 focus:border-brand-500 outline-none"
                  placeholder="Paste destination token..."
                  value={migrateModal.targetToken} onChange={e => setMigrateModal({ ...migrateModal, targetToken: e.target.value })}
                />
              </div>
            </div>

            <div className="flex justify-end gap-3">
              <button onClick={() => setMigrateModal({ ...migrateModal, isOpen: false })} className="px-5 py-2.5 text-slate-400 hover:text-white font-bold text-sm">Cancel</button>
              <button onClick={handleMigrate} className="px-5 py-2.5 bg-brand-600 hover:bg-brand-500 text-white rounded-xl font-bold text-sm shadow-[0_0_15px_#0ea5e950]">Initiate Transfer</button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}