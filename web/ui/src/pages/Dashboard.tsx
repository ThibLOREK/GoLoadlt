import { useQuery } from "@tanstack/react-query";
import { Box, Card, CardContent, Typography, Grid, LinearProgress, Chip, Stack } from "@mui/material";
import { api, Run } from "../api/client";

function StatCard({ label, value, color }: { label: string; value: string | number; color?: string }) {
  return (
    <Card>
      <CardContent>
        <Typography variant="caption" color="text.secondary">{label}</Typography>
        <Typography variant="h4" fontWeight="bold" color={color}>{value}</Typography>
      </CardContent>
    </Card>
  );
}

export default function Dashboard() {
  const { data: pipelines = [] } = useQuery({
    queryKey: ["pipelines"],
    queryFn: () => api.get("/pipelines").then(r => r.data),
    refetchInterval: 5000,
  });

  const allRunsQueries = pipelines.slice(0, 10).map((p: { id: string }) => p.id);

  const { data: recentRuns = [] } = useQuery<Run[]>({
    queryKey: ["dashboard-runs", allRunsQueries],
    queryFn: async () => {
      const all: Run[] = [];
      for (const id of allRunsQueries) {
        const runs = await api.get<Run[]>(`/pipelines/${id}/runs`).then(r => r.data);
        all.push(...runs.slice(0, 3));
      }
      return all.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).slice(0, 20);
    },
    refetchInterval: 5000,
    enabled: pipelines.length > 0,
  });

  const stats = {
    total: pipelines.length,
    running: recentRuns.filter((r: Run) => r.status === "running").length,
    succeeded: recentRuns.filter((r: Run) => r.status === "succeeded").length,
    failed: recentRuns.filter((r: Run) => r.status === "failed").length,
  };

  const STATUS_COLOR: Record<string, string> = {
    pending: "#9e9e9e", running: "#ff9800", succeeded: "#4caf50", failed: "#f44336", cancelled: "#9c27b0",
  };

  return (
    <Box>
      <Typography variant="h5" fontWeight="bold" mb={3}>Dashboard</Typography>

      <Grid container spacing={2} mb={4}>
        <Grid item xs={3}><StatCard label="Pipelines" value={stats.total} /></Grid>
        <Grid item xs={3}><StatCard label="En cours" value={stats.running} color="#ff9800" /></Grid>
        <Grid item xs={3}><StatCard label="Succès (récents)" value={stats.succeeded} color="#4caf50" /></Grid>
        <Grid item xs={3}><StatCard label="Échecs (récents)" value={stats.failed} color="#f44336" /></Grid>
      </Grid>

      <Typography variant="h6" mb={2}>Activité récente</Typography>
      <Stack spacing={1}>
        {recentRuns.map((run: Run) => (
          <Card key={run.id} sx={{ borderLeft: `4px solid ${STATUS_COLOR[run.status] ?? "#666"}` }}>
            <CardContent sx={{ py: 1, "&:last-child": { pb: 1 } }}>
              <Stack direction="row" justifyContent="space-between" alignItems="center">
                <Stack>
                  <Typography variant="caption" fontFamily="monospace">{run.pipeline_id.slice(0, 8)}…</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {new Date(run.created_at).toLocaleString()}
                  </Typography>
                </Stack>
                <Stack direction="row" spacing={1} alignItems="center">
                  <Typography variant="caption">{run.records_loaded} chargés</Typography>
                  <Chip label={run.status} size="small"
                    sx={{ bgcolor: STATUS_COLOR[run.status], color: "white" }} />
                </Stack>
              </Stack>
              {run.status === "running" && <LinearProgress sx={{ mt: 1 }} />}
            </CardContent>
          </Card>
        ))}
      </Stack>
    </Box>
  );
}
