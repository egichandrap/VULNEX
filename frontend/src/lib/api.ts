// VULNEX Typed API Client

export interface APIError {
  code: string;
  message: string;
}

export interface SetupAdminRequest {
  setup_token: string;
  email: string;
  password: string;
  display_name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface WorkspaceDTO {
  id: string;
  name: string;
  owner_user_id: string;
  allow_private_network_scan: boolean;
  status: string;
}

export interface AssetDTO {
  id: string;
  project_id: string;
  asset_type: string;
  canonical_value: string;
  display_value: string;
  criticality: string;
  status: string;
}

export interface ScanDTO {
  id: string;
  workspace_id: string;
  project_id: string;
  status: string;
  targets: string[];
  started_at: string | null;
  completed_at: string | null;
  failure_reason: string;
  created_at: string;
}

export interface FindingDTO {
  id: string;
  workspace_id: string;
  project_id: string;
  scan_id: string;
  asset_id: string;
  fingerprint: string;
  title: string;
  description: string;
  severity: "INFO" | "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
  confidence: "LOW" | "MEDIUM" | "HIGH" | "CONFIRMED";
  status: "OPEN" | "IN_PROGRESS" | "RESOLVED" | "ACCEPTED_RISK" | "FALSE_POSITIVE";
  source_scanner: string;
  first_seen_at: string;
  last_seen_at: string;
  risk_score: number;
}

class APIClient {
  private baseURL: string;
  private token: string | null = null;

  constructor() {
    this.baseURL = typeof window !== 'undefined'
      ? ''
      : (process.env.INTERNAL_API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://api:8080');
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('vulnex_token');
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('vulnex_token', token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('vulnex_token');
    }
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers,
    });

    const data = await response.json().catch(() => null);

    if (!response.ok) {
      const err = data?.error || { code: 'UNKNOWN_ERROR', message: `HTTP ${response.status}` };
      throw new Error(err.message || 'Request failed');
    }

    return data as T;
  }

  // Auth endpoints
  async setupInitialAdmin(req: SetupAdminRequest) {
    return this.request<{ user_id: string; email: string }>('/api/v1/auth/setup', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  }

  async login(req: LoginRequest) {
    const res = await this.request<{ token: string; user: { id: string; email: string; display_name: string } }>(
      '/api/v1/auth/login',
      {
        method: 'POST',
        body: JSON.stringify(req),
      }
    );
    this.setToken(res.token);
    return res;
  }

  async getMe() {
    return this.request<{ id: string; email: string; display_name: string; status: string }>('/api/v1/auth/me');
  }

  // Workspace endpoints
  async listWorkspaces() {
    return this.request<{ workspaces: WorkspaceDTO[] }>('/api/v1/workspaces');
  }

  async createWorkspace(name: string) {
    return this.request<{ workspace_id: string; name: string }>('/api/v1/workspaces', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  }

  async updateWorkspacePolicy(workspaceId: string, allowPrivate: boolean, justification: string) {
    return this.request<{ status: string }>(`/api/v1/workspaces/${workspaceId}/network-policy`, {
      method: 'PATCH',
      body: JSON.stringify({
        allow_private_network_scan: allowPrivate,
        justification,
      }),
    });
  }

  // Asset endpoints
  async listAssets(projectId: string) {
    return this.request<{ assets: AssetDTO[] }>(`/api/v1/projects/${projectId}/assets`);
  }

  async createAsset(projectId: string, asset: { asset_type: string; raw_value: string; display_value: string; criticality?: string }) {
    return this.request<{ id: string; canonical_value: string; status: string }>(`/api/v1/projects/${projectId}/assets`, {
      method: 'POST',
      body: JSON.stringify(asset),
    });
  }

  // Scan endpoints
  async requestScan(projectId: string, payload: {
    workspace_id: string;
    authorization_id: string;
    profile_version_id: string;
    idempotency_key: string;
    targets: string[];
  }) {
    return this.request<{ scan_id: string; status: string }>(`/api/v1/projects/${projectId}/scans`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  }

  async getScan(scanId: string) {
    return this.request<ScanDTO>(`/api/v1/scans/${scanId}`);
  }

  async cancelScan(scanId: string, reason: string) {
    return this.request<{ status: string }>(`/api/v1/scans/${scanId}/cancel`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    });
  }

  // Finding endpoints
  async listFindings(projectId: string) {
    return this.request<{ findings: FindingDTO[] }>(`/api/v1/projects/${projectId}/findings`);
  }

  async triageFinding(findingId: string, status: string, justification: string) {
    return this.request<{ id: string; status: string }>(`/api/v1/findings/${findingId}`, {
      method: 'PATCH',
      body: JSON.stringify({ status, justification }),
    });
  }
}

export const api = new APIClient();
