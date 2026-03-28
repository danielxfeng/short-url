import config from '@/config/config';
import { HealthCheckResSchema } from 'schemas';

const healthCheck = async () => {
  const response = await fetch(`${config.apiBaseUrl}/health`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to perform health check: ${response.statusText}`);
  }

  return HealthCheckResSchema.parse(await response.json());
};

export default { healthCheck };
