import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

export default function AuthCallback() {
  const { provider } = useParams();
  const navigate = useNavigate();

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');

    if (!token) {
      console.error('No token received');
      navigate('/');
      return;
    }

    localStorage.setItem('aeshield_token', token);
    navigate('/dashboard', { replace: true });
  }, [provider, navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-background-light dark:bg-background-dark">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
        <p className="text-charcoal dark:text-slate-100">Đang xác thực...</p>
      </div>
    </div>
  );
}
