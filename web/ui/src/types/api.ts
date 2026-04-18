// Types correspondant aux structs Go (contracts.Project, connections.Connection...)

export interface Param {
  name: string
  value: string
}

export interface ETLNode {
  id: string
  type: string
  label?: string
  connectionRef?: string
  posX?: number
  posY?: number
  params?: Param[]
}

export interface ETLEdge {
  from: string
  to: string
  fromPort?: string
  toPort?: string
}

export interface Project {
  id: string
  name: string
  description?: string
  version: number
  activeEnv?: string
  nodes: ETLNode[]
  edges: ETLEdge[]
}

export interface ConnEnv {
  name: string
  host: string
  port: number
  database: string
  user: string
  secretRef: string
}

export interface Connection {
  id: string
  name: string
  type: string // postgres | mysql | mssql | rest
  envs: Record<string, ConnEnv>
}

export interface BlockMeta {
  type: string
  category: string
  label: string
  description: string
  minInputs: number
  maxInputs: number
  minOutputs: number
  maxOutputs: number
}

export interface RunResult {
  nodeID: string
  rowsIn: number
  rowsOut: number
  duration: string
  err?: string
}

export interface RunReport {
  projectID: string
  success: boolean
  duration: string
  results: RunResult[]
}
