import { expect, test } from 'vitest';
import { HealthCheckResSchema } from './schemas';

test('validates HealthCheckResSchema', () => {
  const testCases = [
    { input: { status: 'ok' }, expected: true },
    { input: {}, expected: false },
  ];

  testCases.forEach(({ input, expected }) => {
    const result = HealthCheckResSchema.safeParse(input);
    expect(result.success).toBe(expected);
  });
});
