import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Files from './pages/Files';
import Settings from './pages/Settings';
import AuthCallback from './pages/AuthCallback';
import { NotificationsProvider } from './context/NotificationsContext';

function ProtectedRoute({ children }) {
  const token = localStorage.getItem('aeshield_token');

  if (!token) {
    return <Navigate to="/" replace />;
  }

  return <NotificationsProvider>{children}</NotificationsProvider>;
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