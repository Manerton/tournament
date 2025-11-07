import React from 'react';
import LiveTable from "./TableTemplate";
import {useParams} from "react-router-dom";
import {API_URL} from "../../config/routes.ts";
import type {LeaderboardWinDTOResponse} from "../../types/LeaderboadTypes.ts";

const WinLeaderboard: React.FC = () => {
    const { id } = useParams<{ id: string }>();

    return (
        <div className="container-fluid">
            <h1 className="text-2xl font-bold mb-4"></h1>
            <LiveTable<LeaderboardWinDTOResponse>
                title="Таблица по победам"
                url={`${API_URL}/api/leaderboard/wins/${id}`}
                idKey="id"
                sortKey="scoreWins"
                sortKey2="score"
                columns={[
                    /*{ key: "id", label: "ID" },*/
                    { key: "name", label: "Название команды" },
                    { key: "scoreWins", label: "Победы" },
                    { key: "score", label: "Очки" },
                ]}
            />
        </div>
    );
};

export default WinLeaderboard;