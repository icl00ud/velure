export interface ILoginUser {
    email: string;
    password: string;
}

export interface ILoginResponse {
    accessToken: string,
    refreshToken: string
}

export interface IRegisterUser { 
    name: string;
    email: string;
    password: string;
}

type User = {
    id: number;
    name: string;
    email: string;
    createdAt: Date;
    updatedAt: Date;
}