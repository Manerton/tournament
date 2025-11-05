import React from "react";
import { useParams } from "react-router-dom";
import LiveTable from "./TableTemplate.tsx";
import { API_URL } from "../../config/routes.ts";
import { Col, Row } from "react-bootstrap";
import type {LeaderboardPlayersDTOResponse, LeaderboardDTOResponse, LeaderboardWinDTOResponse} from "../../types/LeaderboadTypes.ts";

const GlobalLeaderboard: React.FC = () => {
    const { id } = useParams<{ id: string }>();

    return (
        <div className="container-fluid">
            <Row>
                <Col>
                    <div className="table-responsive">
                        <LiveTable<LeaderboardPlayersDTOResponse>
                            title="Рейтинг игроков турнира"
                            url={`${API_URL}/api/leaderboard/participants/${id}`}
                            idKey="name"
                            sortKey="score"
                            columns={[
                                { key: "team_name", label: "Название команды" },
                                { key: "name", label: "Имя" },
                                { key: "score", label: "Очки" },
                            ]}
                        />
                    </div>
                </Col>

                <Col>
                    <LiveTable<LeaderboardDTOResponse>
                        title="Таблица команд по очкам"
                        url={`${API_URL}/api/leaderboard/score/${id}`}
                        idKey="id"
                        sortKey="score"
                        columns={[
                            { key: "name", label: "Имя" },
                            { key: "score", label: "Очки" },
                        ]}
                    />
                </Col>

                <Col>
                    <LiveTable<LeaderboardWinDTOResponse>
                        title="Таблица по победам"
                        url={`${API_URL}/api/leaderboard/wins/${id}`}
                        idKey="id"
                        sortKey="scoreWins"
                        columns={[
                            { key: "name", label: "Имя" },
                            { key: "scoreWins", label: "Победы" },
                            { key: "score", label: "Очки" },
                        ]}
                    />
                </Col>
            </Row>
        </div>
    );
};

export default GlobalLeaderboard;
