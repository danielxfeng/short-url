import * as z from 'zod';
import { HealthCheckResSchema } from './schemas';

export type HealthCheckRes = z.infer<typeof HealthCheckResSchema>;
