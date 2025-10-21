import { useEffect, useState } from "react"
import { Alert, Card, Container } from "react-bootstrap"
import { axiosGetAllTournaments } from "../../requests/tournamentRequests"
import { EventStatus, type EventResponse } from "../../types/TournamentTypes"

const Tournaments: React.FC = () => {

    const [notify, setNotify] = useState<{ variant: "success" | "danger"; text: string } | null>(null);
    const [tournaments, setTournaments] = useState<EventResponse[]>([])

    useEffect(() => {
        const getTournaments = async () => {
            try {
                const res = await axiosGetAllTournaments()
                setTournaments(res)
            } catch (err) {
                const message = err instanceof Error ? err.message : "Не удалось загрузить муниципалитеты";
                setNotify({ variant: "danger", text: message });
                console.error("Ошибка при загрузке муниципалитет:", err);
            }
        }
        getTournaments()
    }, [])


    return (
        <Container>
            <h2 className="mb-4 fw-bold text-start">Список всех турниров</h2>

            {notify && <Alert variant={notify.variant}>{notify.text}</Alert>}

            {tournaments && tournaments.map((tournament) => (
                <Card key={tournament.id} className="rounded-3 mb-3 bg-transparent glass-card">
                    <div className="d-flex justify-content-between">
                        <div className="ms-3 text-start">
                            <h5>{tournament.name}</h5>

                            <p>Статус: {tournament.status === EventStatus.Active ? "Активен" : "Завершён"}</p>
                        </div>
                        <div className="me-3 d-flex gap-3">
                            <button className="btn btn-outline-primary">Редактировать</button>
                            <button className="btn btn-outline-danger">Удалить</button>
                        </div>
                    </div>
                </Card>
            ))}

        </Container>

    )
}

export default Tournaments