import React from "react";
import ReactLoading from "react-loading";

export default function Loading() {
    return (
        <div>
            <ReactLoading type="bubbles" color="#6c757d"
                height={100} width={50} />
        </div>
    );
}