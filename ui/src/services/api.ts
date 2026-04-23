import type { HostConfig, DockerContainer } from '../types/docker';

// 🛡️ ENTERPRISE GUARD 1: Smart Protocol Resolution
const buildUrl = (hostIp: string, path: string): string => {
  // 1. Strip any existing protocols and trailing slashes the user might have pasted
  let cleanIp = hostIp.replace(/^(https?:\/\/|wss?:\/\/)/, '').replace(/\/$/, '');
  
  // 2. Determine secure context. 
  // If the browser is on HTTPS (e.g. Codespaces), we MUST use HTTPS for the API.
  // Exception: If the target is explicitly 'localhost', allow HTTP.
  const isSecureContext = window.location.protocol === 'https:';
  const isLocalhost = cleanIp.startsWith('localhost') || cleanIp.startsWith('127.0.0.1');
  
  const protocol = (isSecureContext && !isLocalhost) ? 'https' : 'http';
  
  return `${protocol}://${cleanIp}${path}`;
};

export const DockerAPI = {
  getContainers: async (host: HostConfig): Promise<DockerContainer[]> => {
    const res = await fetch(buildUrl(host.ip, '/containers'), {
      headers: { 'X-Auth-Token': host.token }
    });
    if (!res.ok) throw new Error('Unauthorized or Offline');
    return res.json();
  },

  containerAction: async (host: HostConfig, action: 'start' | 'stop' | 'delete', ids: string[], force: boolean = false) => {
    const res = await fetch(buildUrl(host.ip, '/containers/action'), {
      method: 'POST',
      headers: { 'X-Auth-Token': host.token, 'Content-Type': 'application/json' },
      body: JSON.stringify({ action, container_ids: ids, force })
    });
    if (!res.ok) throw new Error(`Failed to ${action} containers`);
    return res.json();
  },

  renameContainer: async (host: HostConfig, id: string, newName: string) => {
    const res = await fetch(buildUrl(host.ip, '/containers/rename'), {
      method: 'POST',
      headers: { 'X-Auth-Token': host.token, 'Content-Type': 'application/json' },
      body: JSON.stringify({ container_id: id, new_name: newName })
    });
    if (!res.ok) throw new Error('Failed to rename container');
    return res.json();
  },

  migrateBatch: async (from: HostConfig, targetIp: string, targetToken: string, ids: string[]) => {
    // Note: targetIp needs to be passed clean so the Go backend can use it
    let cleanTargetIp = targetIp.replace(/^(https?:\/\/|wss?:\/\/)/, '').replace(/\/$/, '');

    const res = await fetch(buildUrl(from.ip, '/migrate-batch'), {
      method: 'POST',
      headers: { 'X-Auth-Token': from.token, 'Content-Type': 'application/json' },
      body: JSON.stringify({
        container_ids: ids,
        remote_addr: cleanTargetIp,
        remote_token: targetToken,
        concurrency: 3
      })
    });
    if (!res.ok) throw new Error('Migration failed to start');
    return res.json();
  }
};