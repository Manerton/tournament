import React from 'react';
import AdminSidebar from './AdminSidebar.tsx';
import { Container } from 'react-bootstrap';
import { Outlet } from 'react-router-dom';
import BackgroundParticles from '../common/BackgroundParticles.tsx';

const AdminLayout: React.FC = () => {
  return (
    <div className="d-flex min-vh-100 position-relative">
      <AdminSidebar />

      {/* Частицы помещаем ВНУТРЬ рабочей области */}
      <div className="flex-grow-1 d-flex flex-column position-relative" style={{ overflow: "hidden" }}>
        <BackgroundParticles />

        <Container fluid="lg" className="flex-grow-1 position-relative" style={{ zIndex: 1 }}>
          <main className="py-4">
            <Outlet />
          </main>
        </Container>
      </div>
    </div>
  );
};

export default AdminLayout;
