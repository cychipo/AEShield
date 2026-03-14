import { useState, useEffect } from 'react';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:6868/api/v1';

export default function Dashboard() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchUser();
  }, []);

  const fetchUser = async () => {
    const token = localStorage.getItem('aeshield_token');
    
    if (!token) {
      window.location.href = '/';
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const userData = await response.json();
        setUser(userData);
      } else {
        localStorage.removeItem('aeshield_token');
        window.location.href = '/';
      }
    } catch (error) {
      console.error('Error fetching user:', error);
      localStorage.removeItem('aeshield_token');
      window.location.href = '/';
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('aeshield_token');
    window.location.href = '/';
  };

  if (loading) {
    return (
      <div className="dashboard-loading">
        <div className="spinner"></div>
        <p>Đang tải...</p>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <h1>AEShield Dashboard</h1>
        <button onClick={handleLogout} className="logout-btn">
          Đăng xuất
        </button>
      </header>
      
      <main className="dashboard-content">
        {user && (
          <div className="user-info">
            <img 
              src={user.avatar || '/default-avatar.png'} 
              alt={user.name} 
              className="user-avatar"
            />
            <h2>Xin chào, {user.name}</h2>
            <p>{user.email}</p>
          </div>
        )}
      </main>
    </div>
  );
}
