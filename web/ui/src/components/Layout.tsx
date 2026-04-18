import { Box, AppBar, Toolbar, Typography, Drawer, List, ListItemButton, ListItemText, Divider } from "@mui/material";
import { useNavigate, useLocation } from "react-router-dom";

const DRAWER_WIDTH = 220;

const NAV = [
  { label: "Dashboard", path: "/" },
  { label: "Pipelines", path: "/pipelines" },
];

export default function Layout({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <Box sx={{ display: "flex" }}>
      <AppBar position="fixed" sx={{ zIndex: 1300 }}>
        <Toolbar>
          <Typography variant="h6" fontWeight="bold">⚡ Go ETL Studio</Typography>
        </Toolbar>
      </AppBar>
      <Drawer variant="permanent" sx={{ width: DRAWER_WIDTH, "& .MuiDrawer-paper": { width: DRAWER_WIDTH, mt: "64px" } }}>
        <List>
          {NAV.map(item => (
            <ListItemButton key={item.path}
              selected={location.pathname === item.path}
              onClick={() => navigate(item.path)}>
              <ListItemText primary={item.label} />
            </ListItemButton>
          ))}
        </List>
        <Divider />
      </Drawer>
      <Box component="main" sx={{ flexGrow: 1, p: 3, mt: "64px", ml: `${DRAWER_WIDTH}px` }}>
        {children}
      </Box>
    </Box>
  );
}
