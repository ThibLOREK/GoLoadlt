import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { Box, Typography, Table, TableBody, TableCell, TableHead, TableRow, Chip, Paper } from "@mui/material";
import { pipelinesApi, Run } from "../api/client";

const STATUS_COLOR: Record<string, "default"|"warning"|"success"|"error"> = {
  pending: "default",
  running: "warning",
  succeeded: "success",
  failed: "error",
};

export default function RunHistory() {
  const { id } = useParams<{ id: string }>();
  const { data: runs = [] } = useQuery({
    queryKey: ["runs", id],
    queryFn: () => pipelinesApi.runs(id!),
    refetchInterval: 3000,
  });

  return (
    <Box>
      <Typography variant="h5" fontWeight="bold" mb={3}>Historique des runs</Typography>
      <Paper>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>Statut</TableCell>
              <TableCell>Début</TableCell>
              <TableCell>Fin</TableCell>
              <TableCell>Lus</TableCell>
              <TableCell>Chargés</TableCell>
              <TableCell>Erreur</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {runs.map((run: Run) => (
              <TableRow key={run.id}>
                <TableCell sx={{ fontFamily: "monospace", fontSize: 12 }}>{run.id.slice(0, 8)}</TableCell>
                <TableCell>
                  <Chip label={run.status} size="small" color={STATUS_COLOR[run.status] ?? "default"} />
                </TableCell>
                <TableCell>{run.started_at ? new Date(run.started_at).toLocaleString() : "—"}</TableCell>
                <TableCell>{run.finished_at ? new Date(run.finished_at).toLocaleString() : "—"}</TableCell>
                <TableCell>{run.records_read}</TableCell>
                <TableCell>{run.records_loaded}</TableCell>
                <TableCell sx={{ color: "error.main", fontSize: 12 }}>{run.error_msg || "—"}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Paper>
    </Box>
  );
}
