import service from '@/services';
import { useEffect, useState } from 'react';

const HomePage = () => {
  const [status, setStatus] = useState('Checking health...');

  useEffect(() => {
    const checkHealth = async () => {
      try {
        const res = await service.healthCheck.healthCheck();
        console.log('Health Check Response:', res);
        setStatus(res.status);
      } catch (error) {
        console.error('Health Check Failed:', error);
        setStatus('failed to check health');
      }
    };

    checkHealth();
  }, []);

  return (
    <div>
      <h1>Home</h1>
      <p>Health Status: {status}</p>
    </div>
  );
};

export default HomePage;
