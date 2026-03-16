import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Files from './pages/Files';
import Whitelist from './pages/Whitelist';
import Settings from './pages/Settings';
import AuthCallback from './pages/AuthCallback';

function ProtectedRoute({ children }) {
  const token = localStorage.getItem('aeshield_token');

  if (!token) {
    return <Navigate to="/" replace />;
  }

  return children;
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        />
        <Route
          path="/files"
          element={
            <ProtectedRoute>
              <Files />
            </ProtectedRoute>
          }
        />
        <Route
          path="/whitelist"
          element={
            <ProtectedRoute>
              <Whitelist />
            </ProtectedRoute>
          }
        />
        <Route
          path="/settings"
          element={
            <ProtectedRoute>
              <Settings />
            </ProtectedRoute>
          }
        />
        <Route path="/auth/:provider/callback" element={<AuthCallback />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;