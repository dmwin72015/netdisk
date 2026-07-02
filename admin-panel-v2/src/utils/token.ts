export function getToken(): string | null {
  return localStorage.getItem('nd.access') || sessionStorage.getItem('nd.access');
}

export function setToken(token: string, autoLogin: boolean): void {
  if (autoLogin) {
    localStorage.setItem('nd.access', token);
    sessionStorage.removeItem('nd.access');
  } else {
    sessionStorage.setItem('nd.access', token);
    localStorage.removeItem('nd.access');
  }
}

export function removeToken(): void {
  localStorage.removeItem('nd.access');
  sessionStorage.removeItem('nd.access');
}
