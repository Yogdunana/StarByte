export function isRegisteredOnly(roles?: string[] | null): boolean {
  return !roles?.some((role) => role !== 'user');
}
export function registeredPageAllowed(path: string): boolean {
  return (
    [
      '/user/profile',
      '/user/settings',
      '/member/applications',
      '/member/application',
      '/announcement',
      '/403',
    ].includes(path) || path.startsWith('/announcement/')
  );
}
