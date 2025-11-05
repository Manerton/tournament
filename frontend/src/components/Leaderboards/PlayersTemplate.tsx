// src/pages/PlayersTemplate.tsx
import React from 'react';
import LiveTable from "./TableTemplate.tsx";
import {useParams} from "react-router-dom";
import {API_URL} from "../../config/routes.ts";
import type { LeaderboardPlayersDTOResponse } from "../../types/LeaderboadTypes.ts";

const PlayersTemplate: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    return (
        <div className="container-fluid">
            <div className="table-responsive">
                <LiveTable<LeaderboardPlayersDTOResponse>
                    title="Рейтинг игроков турнира"
                    url={`${API_URL}/api/leaderboard/participants/${id}`}
                    idKey="name" // Используем name как ключ, если ID нет
                    sortKey="score"
                    columns={[
                        { key: "team_name", label: "Название команды" },
                        { key: "name", label: "Имя" },
                        { key: "score", label: "Очки" },
                    ]}
                />
            </div>
        </div>
    );
};

export default PlayersTemplate;