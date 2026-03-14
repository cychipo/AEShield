import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';

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
      </Routes>
    </BrowserRouter>
  );
}

export default App;
