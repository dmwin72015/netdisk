import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router';
import { message } from 'antd';

interface StoredUser {
  role: string;
  username: string;
  [key: string]: unknown;
}

export function useAuthGuard() {
  const navigate = useNavigate();
  const [checking, setChecking] = useState(true);
  const [authorized, setAuthorized] = useState(false);

  useEffect(() => {
    try {
      const raw = localStorage.getItem('admin.user');
      if (!raw) {
        navigate('/login');
        return;
      }
      const user = JSON.parse(raw) as StoredUser;
      if (!user.role || user.role !== 'admin') {
        message.error('仅管理员可访问');
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

export function getStoredUser(): StoredUser | null {
  try {
    const raw = localStorage.getItem('admin.user');
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function logout() {
  localStorage.removeItem('admin.user');
  localStorage.removeItem('admin.token');
}
