import axios from "axios";

export const api = axios.create({ baseURL: "/api/v1" });

export interface Pipeline {
  id: string;
  name: string;
  description: string;
  status: string;
  source_type: string;
  target_type: string;
  source_config: unknown;
  target_config: unknown;
  steps: unknown[];
  created_at: string;
  updated_at: string;
}

export interface Run {
  id: string;
  pipeline_id: string;
  status: string;
  started_at: string | null;
  finished_at: string | null;
  error_msg: string;
  records_read: number;
  records_loaded: number;
  created_at: string;
}

export const pipelinesApi = {
  list: () => api.get<Pipeline[]>("/pipelines").then(r => r.data),
  get: (id: string) => api.get<Pipeline>(`/pipelines/${id}`).then(r => r.data),
  create: (data: Partial<Pipeline>) => api.post<Pipeline>("/pipelines", data).then(r => r.data),
  update: (id: string, data: Partial<Pipeline>) => api.put<Pipeline>(`/pipelines/${id}`, data).then(r => r.data),
  delete: (id: string) => api.delete(`/pipelines/${id}`),
  run: (id: string) => api.post<Run>(`/pipelines/${id}/runs`).then(r => r.data),
  runs: (id: string) => api.get<Run[]>(`/pipelines/${id}/runs`).then(r => r.data),
};
