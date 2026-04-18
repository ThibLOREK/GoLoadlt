import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Box, Button, Card, CardContent, TextField, Typography, Alert } from "@mui/material";
import { api } from "../api/client";

export default function Login() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("admin@etl.local");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const handleLogin = async () => {
    try {
      const resp = await api.post<{ token: string }>("/auth/login", { email, password });
      localStorage.setItem("token", resp.data.token);
      api.defaults.headers.common["Authorization"] = `Bearer ${resp.data.token}`;
      navigate("/");
    } catch {
      setError("Email ou mot de passe invalide");
    }
  };

  return (
    <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh">
      <Card sx={{ width: 400 }}>
        <CardContent>
          <Typography variant="h5" fontWeight="bold" mb={3}>Go ETL Studio</Typography>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
          <TextField
            fullWidth label="Email" value={email}
            onChange={e => setEmail(e.target.value)} sx={{ mb: 2 }}
          />
          <TextField
            fullWidth label="Mot de passe" type="password" value={password}
            onChange={e => setPassword(e.target.value)}
            onKeyDown={e => e.key === "Enter" && handleLogin()}
            sx={{ mb: 3 }}
          />
          <Button fullWidth variant="contained" onClick={handleLogin}>
            Se connecter
          </Button>
        </CardContent>
      </Card>
    </Box>
  );
}
