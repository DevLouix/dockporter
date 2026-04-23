export type MigrationStatus = 
  | 'Pending' 
  | 'Compressing' 
  | 'Sending' 
  | 'Extracting' 
  | 'Success' 
  | 'Failed';

export interface MigrationEvent {
  container_id: string;
  status: MigrationStatus;
  progress: number;
  error?: string;
  timestamp: string;
}

export interface DockerContainer {
  Id: string;
  Names: string[];
  Image: string;
  State: 'running' | 'exited' | 'paused';
  Status: string;
}

export interface HostConfig {
  ip: string;
  token: string;
  nickname: string;
}