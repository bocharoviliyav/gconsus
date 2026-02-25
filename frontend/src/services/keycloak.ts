import Keycloak from 'keycloak-js';
import { config } from '../config';

let keycloakInstance: Keycloak | null = null;

export const initKeycloak = (): Keycloak => {
  if (keycloakInstance) {
    return keycloakInstance;
  }

  keycloakInstance = new Keycloak({
    url: config.keycloak.url,
    realm: config.keycloak.realm,
    clientId: config.keycloak.clientId,
  });

  return keycloakInstance;
};

export const getKeycloak = (): Keycloak => {
  if (!keycloakInstance) {
    throw new Error('Keycloak not initialized. Call initKeycloak() first.');
  }
  return keycloakInstance;
};

export const login = async (): Promise<boolean> => {
  const keycloak = getKeycloak();
  try {
    const authenticated = await keycloak.init({
      onLoad: 'login-required',
      checkLoginIframe: false,
      pkceMethod: 'S256',
    });
    return authenticated;
  } catch (error) {
    console.error('Failed to initialize Keycloak', error);
    return false;
  }
};

export const logout = (): void => {
  const keycloak = getKeycloak();
  keycloak.logout({
    redirectUri: window.location.origin,
  });
};

export const getToken = (): string | undefined => {
  const keycloak = getKeycloak();
  return keycloak.token;
};

export const updateToken = async (minValidity: number = 70): Promise<string | undefined> => {
  const keycloak = getKeycloak();
  try {
    const refreshed = await keycloak.updateToken(minValidity);
    if (refreshed) {
      console.log('Token refreshed');
    }
    return keycloak.token;
  } catch (error) {
    console.error('Failed to refresh token', error);
    logout();
    return undefined;
  }
};

export const getUserRoles = (): string[] => {
  const keycloak = getKeycloak();
  if (!keycloak.realmAccess) {
    return [];
  }
  return keycloak.realmAccess.roles || [];
};

export const hasRole = (role: string): boolean => {
  const keycloak = getKeycloak();
  return keycloak.hasRealmRole(role);
};

export const hasAnyRole = (roles: string[]): boolean => {
  return roles.some((role) => hasRole(role));
};

export const getUserInfo = () => {
  const keycloak = getKeycloak();
  if (!keycloak.tokenParsed) {
    return null;
  }

  return {
    sub: keycloak.tokenParsed.sub || '',
    email: keycloak.tokenParsed.email || '',
    preferred_username: keycloak.tokenParsed.preferred_username || '',
    name: keycloak.tokenParsed.name || '',
    given_name: keycloak.tokenParsed.given_name,
    family_name: keycloak.tokenParsed.family_name,
    roles: getUserRoles(),
  };
};

export const isAuthenticated = (): boolean => {
  const keycloak = getKeycloak();
  return keycloak.authenticated || false;
};
