import type { CreateActivityParams, UpdateActivityParams } from '@/api/activity';

function isFilled(value: unknown): boolean {
  return value !== undefined && value !== null && value !== '';
}

export function toCreateParams(values: Record<string, unknown>): CreateActivityParams {
  return {
    title: String(values.title),
    description: values.description ? String(values.description) : undefined,
    category: values.category ? String(values.category) : undefined,
    tags: Array.isArray(values.tags) ? (values.tags as string[]) : undefined,
    location: values.location ? String(values.location) : undefined,
    max_participants: isFilled(values.max_participants) ? Number(values.max_participants) : undefined,
    latitude: isFilled(values.latitude) ? Number(values.latitude) : undefined,
    longitude: isFilled(values.longitude) ? Number(values.longitude) : undefined,
    checkin_radius_m: isFilled(values.checkin_radius_m) ? Number(values.checkin_radius_m) : undefined,
    start_time: String(values.start_time),
    end_time: String(values.end_time),
  };
}

export function toUpdateParams(values: Record<string, unknown>): UpdateActivityParams {
  const params = toCreateParams(values);
  const noGeo = !isFilled(values.latitude) && !isFilled(values.longitude) && !isFilled(values.checkin_radius_m);
  return { ...params, clear_geo: noGeo };
}
