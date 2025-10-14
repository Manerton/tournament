import React from 'react';
import { Container } from 'react-bootstrap';
import { Outlet } from 'react-router-dom';

const RegisterFormLayout: React.FC = () => {
  return (
    <>
      <div className="d-flex flex-column min-vh-100">
        <Container fluid="lg" className="flex-grow-1">
          <main className="py-4">
            <Outlet />
          </main>
        </Container>
      </div>
    </>
  );
};

export default RegisterFormLayout;
