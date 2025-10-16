// src/components/Layout.tsx
import React from "react";
import { Container } from "react-bootstrap";
import Header from "./Header";
import { Outlet } from "react-router-dom";
import BackgroundParticles from "./BackgroundParticles";
import Footer from "./Footer";

const Layout: React.FC = () => {
  return (
    <>
      <Header />
      <BackgroundParticles />
      <div className="d-flex flex-column min-vh-100 position-relative">
        <Container fluid="lg" className="flex-grow-1">
          <main className="py-4 position-relative" >
            <Outlet />
          </main>
        </Container>
        <Footer />
      </div>
    </>
  );
};

export default Layout;
