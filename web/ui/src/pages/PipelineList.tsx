import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { Box, Button, Card, CardContent, Typography, Stack, Chip } from "@mui/material";
import { pipelinesApi } from "../api/client";

const STATUS_COLOR: Record<string, "default"|"warning"|"success"|"error"> = {
  draft: "default",
  ready: "warning",
  running: "warning",
  succeeded: "success",
  failed: "error",
};

export default function PipelineList() {
  const navigate = useNavigate();
  const qc = useQueryClient();
  const { data: pipelines = [] } = useQuery({ queryKey: ["pipelines"], queryFn: pipelinesApi.list });

  const runMutation = useMutation({
    mutationFn: (id: string) => pipelinesApi.run(id),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ["runs", id] });
    },
  });

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" fontWeight="bold">Pipelines</Typography>
        <Button variant="contained" onClick={() => navigate("/pipelines/new/design")}>
          + Nouveau pipeline
        </Button>
      </Stack>
      <Stack spacing={2}>
        {pipelines.map(p => (
          <Card key={p.id}>
            <CardContent>
              <Stack direction="row" justifyContent="space-between" alignItems="center">
                <Box>
                  <Typography variant="h6">{p.name}</Typography>
                  <Typography variant="body2" color="text.secondary">{p.description}</Typography>
                  <Stack direction="row" spacing={1} mt={1}>
                    <Chip label={p.source_type} size="small" />
                    <Chip label="→" size="small" />
                    <Chip label={p.target_type} size="small" />
                    <Chip label={p.status} size="small" color={STATUS_COLOR[p.status] ?? "default"} />
                  </Stack>
                </Box>
                <Stack direction="row" spacing={1}>
                  <Button size="small" onClick={() => navigate(`/pipelines/${p.id}/design`)}>Designer</Button>
                  <Button size="small" onClick={() => navigate(`/pipelines/${p.id}/runs`)}>Runs</Button>
                  <Button size="small" variant="contained" color="success"
                    onClick={() => runMutation.mutate(p.id)}>
                    ▶ Run
                  </Button>
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        ))}
      </Stack>
    </Box>
  );
}
