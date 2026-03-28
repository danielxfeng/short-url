import * as z from 'zod';

export const HealthCheckResSchema = z.object({
  status: z.literal('ok'),
});
