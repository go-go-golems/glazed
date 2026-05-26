import { Route, Routes } from 'react-router-dom';
import App from './App';

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/:package/:version/sections/:slug" element={<App />} />
      <Route path="/:package/:version" element={<App />} />
      <Route path="*" element={<App />} />
    </Routes>
  );
}
