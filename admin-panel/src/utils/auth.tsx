import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router';
import { message } from 'antd';

interface StoredUser {
  role: string;
  [key: string]: unknown;
}

export function useAuthGuard() {
  const navigate = useNavigate();
  const [checking, setChecking] = useState(true);
  const [authorized, setAuthorized] = useState(false);

  useEffect(() => {
    try {
      const raw = localStorage.getItem('nd.user');
      if (!raw) {
        navigate('/login');
        return;
      }
      const user = JSON.parse(raw) as StoredUser;
      if (!user.role || user.role !== 'admin') {
        message.error('Admin only');
        navigate('/login');
        return;
      }
      setAuthorized(true);
    } catch {
      navigate('/login');
    } finally {
      setChecking(false);
    }
  }, [navigate]);

  return { checking, authorized };
}
