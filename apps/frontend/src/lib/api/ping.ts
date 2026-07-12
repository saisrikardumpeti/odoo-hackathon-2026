import axios from 'axios';

export interface Ping {
  message: string
}

export const fetchPing = async (): Promise<Ping> => {
  const { data } = await axios.get('/api/example');
  return data;
};
