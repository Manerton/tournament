import React from 'react';
import LiveTable from "./TableTemplate.tsx";
import {useParams} from "react-router-dom";
import {API_URL} from "../../config/routes.ts";

import type {LeaderboardDTOResponse} from "../../types/LeaderboadTypes.ts";


const ScoreTemplate: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    return (
        <div className="container-fluid">

            <LiveTable<LeaderboardDTOResponse>
                title="Таблица по очкам"
                url={`${API_URL}/api/leaderboard/score/${id}`}
                idKey="id"
                sortKey="score"
                columns={[
                    /*{ key: "id", label: "ID" },*/
                    { key: "name", label: "Имя" },
                    { key: "score", label: "Очки" },
                ]}
            />
        </div>
    );
};

export default ScoreTemplate;