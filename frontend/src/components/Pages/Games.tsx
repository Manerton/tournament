import type React from "react";
import { useEffect, useState } from "react";
import { Alert, Card, Container } from "react-bootstrap";
import { useParams } from "react-router-dom";
import { axiosGetGames } from "../../requests/gamesRequests";
import type { GameAllInformation } from "../../types/gameType";

const GamesPage: React.FC = () => {

    const { id } = useParams<{ id: string }>();

    const [games, setGames] = useState<GameAllInformation[]>([])


    useEffect(() => {
        const loadGames = async () => {
            const res = await axiosGetGames(id!)
            console.log(res)
            setGames(res)
        }
        loadGames();
    }, [])

    return (
        <Container className="py-4">
            <h2 className="mb-4 fw-bold text-center display-6">Список игр</h2>
            {games.length === 0 ? (
                <Alert variant="info">Нет созданных игр</Alert>
            ) : (
                games.map((g) => (
                    <div className="d-flex  justify-content-between m-3 align-items-center">
                        <Card
                            className="rounded-6 shadow-sm liquid-glass-bg card-interactive"
                        >
                            <Card.Body className="py-4">
                                {g.team_a_name}
                            </Card.Body>
                        </Card>
                        VS
                        <Card
                            className="rounded-6 shadow-sm liquid-glass-bg card-interactive"
                        >
                            <Card.Body className="py-4">
                                {g.team_b_name}
                            </Card.Body>
                        </Card>
                    </div>
                ))
            )
            }
        </Container >
    )
}

export default GamesPage
