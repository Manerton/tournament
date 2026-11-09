import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { Spinner } from "react-bootstrap";
import { useAuth } from "./AuthContext";
import { UserRole } from "../../types/userType";

const RequireAdmin: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { user, refresh } = useAuth();
    const [ready, setReady] = useState(false);

    useEffect(() => {
        let cancelled = false;
        refresh().finally(() => {
            if (!cancelled) setReady(true);
        });
        return () => {
            cancelled = true;
        };
    }, []);

    if (!ready) {
        return (
            <div className="d-flex justify-content-center align-items-center min-vh-100">
                <Spinner animation="border" variant="primary" />
            </div>
        );
    }

    if (user?.role !== UserRole.Admin) {
        return <Navigate to="/access-denied" replace />;
    }

    return <>{children}</>;
};

export default RequireAdmin;
