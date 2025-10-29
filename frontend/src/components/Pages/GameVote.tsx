import { Alert, Container, Spinner } from "react-bootstrap";
import type { GameAllInformation } from "../../types/gameType";
import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";


interface Props {
    game: GameAllInformation;
}

const GameVotePage: React.FC = () => {
    const location = useLocation();
    const navigate = useNavigate();

    const [loading, setLoading] = useState(true);
    const { game } = (location.state || {}) as Props;

    useEffect(() => {
        const loadingTeams = async () => {

        }
        loadingTeams();
    }, [])


    return (
        <Container className="py-4">
            <h2 className="mb-4 fw-bold text-center display-6">Оценка игр турнира</h2>
            <div className="d-flex justify-content-center ">
                <h3 className="mb-4 fw-bold text-center display-6">Матч:</h3>
                <h3 className="mb-4 fw-bold text-center display-6">{game.team_a_name}</h3>
                <h3 className="mb-4 fw-bold text-center display-6"> vs </h3>
                <h3 className="mb-4 fw-bold text-center display-6">{game.team_b_name}</h3>
            </div>


            <Container className="py-4">
                <Alert variant="danger" className="mt-4 text-center">
                    Нет данных об игре. Вернуться <Alert.Link onClick={() => navigate(-1)}>назад</Alert.Link>
                </Alert>
            </Container>

            {loading ? (
                <div className="text-center my-5">
                    <Spinner animation="border" />
                </div>
            ) : game ? (
                <Alert variant="info">Нет такой игры</Alert>
            ) : (
                <div className="accardion"></div>
            )}
        </Container>
    );
}

export default GameVotePage;