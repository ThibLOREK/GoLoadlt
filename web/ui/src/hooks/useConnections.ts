import { useEffect, useState } from 'react'
import { listConnections } from '@/api/client'
import type { Connection } from '@/types/api'

export function useConnections(filterType?: string): Connection[] {
  const [connections, setConnections] = useState<Connection[]>([])

  useEffect(() => {
    listConnections().then(all =>
      setConnections(filterType ? all.filter(c => c.type === filterType) : all)
    )
  }, [filterType])

  return connections
}
