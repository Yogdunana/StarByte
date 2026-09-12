export function engineReviewClosed(status: number): boolean {
  return status === 3 || status === 4 || status === 5;
}
