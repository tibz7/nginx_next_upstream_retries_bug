import  http  from "k6/http";

// Options
export const options = {
    stages: [
        { target: 500, duration: "2m" },
    ],
};

export default function () {
    let request = `http://host.docker.internal:2222/ok`
    let response = http.get(request);
   
    // // just for debug in console 
    if (response.status !== 200) {
        console.log(response.status);
        console.log(request);
    }
};
