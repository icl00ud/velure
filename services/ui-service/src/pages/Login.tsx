import { Eye, EyeOff, Heart, Lock, Mail, User } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { useAuth } from "@/hooks/use-auth";
import { toast } from "@/hooks/use-toast";
import { designSystemStyles } from "@/styles/design-system";

const Login = () => {
  const [isLogin, setIsLogin] = useState(true);
  const [showPassword, setShowPassword] = useState(false);
  const [isVisible, setIsVisible] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    password: "",
    confirmPassword: "",
  });

  const { login, register, isLoading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    setIsVisible(true);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      if (isLogin) {
        await login({
          email: formData.email,
          password: formData.password,
        });
        toast({
          title: "Bem-vindo de volta!",
          description: "Você foi autenticado com sucesso.",
        });
        navigate("/");
      } else {
        if (formData.password !== formData.confirmPassword) {
          toast({
            title: "Erro",
            description: "As senhas não coincidem.",
            variant: "destructive",
          });
          return;
        }

        await register({
          name: formData.name,
          email: formData.email,
          password: formData.password,
        });
        toast({
          title: "Conta criada!",
          description: "Bem-vindo ao Velure!",
        });
        navigate("/");
      }
    } catch (error) {
      toast({
        title: "Erro",
        description: error instanceof Error ? error.message : "Ocorreu um erro",
        variant: "destructive",
      });
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-gradient-to-br from-[#F8FAF5] via-[#EDF7ED] to-[#E8F5E9] flex items-center justify-center p-4 grain-texture relative overflow-hidden">
        {/* Decorative Elements */}
        <div className="fixed top-20 right-10 w-64 h-64 rounded-full bg-[#52B788]/10 blur-3xl pointer-events-none" />
        <div className="fixed bottom-20 left-10 w-80 h-80 rounded-full bg-[#95D5B2]/10 blur-3xl pointer-events-none" />

        <div className="w-full max-w-md relative z-10">
          {/* Logo */}
          <div className={`text-center mb-8 ${isVisible ? 'hero-enter active' : 'hero-enter'}`}>
            <Link to="/" className="inline-flex items-center space-x-3 mb-6 group">
              <div className="relative">
                <div className="absolute inset-0 bg-gradient-to-br from-[#52B788] to-[#40916C] rounded-2xl blur-md group-hover:blur-lg transition-all opacity-50" />
                <div className="relative bg-gradient-to-br from-[#52B788] to-[#40916C] rounded-2xl p-3 transform group-hover:scale-110 transition-transform duration-300">
                  <Heart className="h-8 w-8 text-white fill-white" />
                </div>
              </div>
              <span className="font-display font-bold text-3xl text-[#1B4332] group-hover:text-[#52B788] transition-colors">
                Velure
              </span>
            </Link>
            <p className="font-body text-lg text-[#2D6A4F]">
              {isLogin ? "Bem-vindo de volta!" : "Junte-se à família Velure"}
            </p>
          </div>

          <Card className={`shadow-2xl border-2 border-[#1B4332]/10 rounded-3xl backdrop-blur-sm bg-white/95 ${isVisible ? 'page-enter active' : 'page-enter'}`} style={{ animationDelay: '0.2s' }}>
            <CardHeader className="text-center pt-8">
              <CardTitle className="font-display text-4xl font-bold text-[#1B4332]">
                {isLogin ? "Entrar" : "Criar conta"}
              </CardTitle>
              <CardDescription className="font-body text-base text-[#2D6A4F] mt-3">
                {isLogin
                  ? "Digite suas credenciais para acessar sua conta"
                  : "Cadastre-se para começar a comprar para seus pets"}
              </CardDescription>
            </CardHeader>

            <CardContent className="space-y-6 px-8 pb-8">
              <form onSubmit={handleSubmit} className="space-y-5">
                {!isLogin && (
                  <div className="space-y-2">
                    <label htmlFor="name" className="font-body text-sm font-semibold text-[#1B4332]">
                      Nome completo
                    </label>
                    <div className="relative">
                      <User className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-[#2D6A4F]" />
                      <Input
                        id="name"
                        name="name"
                        type="text"
                        placeholder="Digite seu nome completo"
                        value={formData.name}
                        onChange={handleInputChange}
                        required={!isLogin}
                        className="pl-12 font-body border-2 border-[#1B4332]/10 rounded-xl h-12 focus:border-[#52B788]"
                      />
                    </div>
                  </div>
                )}

                <div className="space-y-2">
                  <label htmlFor="email" className="font-body text-sm font-semibold text-[#1B4332]">
                    E-mail
                  </label>
                  <div className="relative">
                    <Mail className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-[#2D6A4F]" />
                    <Input
                      id="email"
                      name="email"
                      type="email"
                      placeholder="Digite seu e-mail"
                      value={formData.email}
                      onChange={handleInputChange}
                      required
                      className="pl-12 font-body border-2 border-[#1B4332]/10 rounded-xl h-12 focus:border-[#52B788]"
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <label htmlFor="password" className="font-body text-sm font-semibold text-[#1B4332]">
                    Senha
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-[#2D6A4F]" />
                    <Input
                      id="password"
                      name="password"
                      type={showPassword ? "text" : "password"}
                      placeholder="Digite sua senha"
                      value={formData.password}
                      onChange={handleInputChange}
                      required
                      className="pl-12 pr-12 font-body border-2 border-[#1B4332]/10 rounded-xl h-12 focus:border-[#52B788]"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="absolute right-2 top-1/2 transform -translate-y-1/2 h-8 w-8 rounded-full hover:bg-[#52B788]/10"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      {showPassword ? (
                        <EyeOff className="h-5 w-5 text-[#2D6A4F]" />
                      ) : (
                        <Eye className="h-5 w-5 text-[#2D6A4F]" />
                      )}
                    </Button>
                  </div>
                </div>

                {!isLogin && (
                  <div className="space-y-2">
                    <label
                      htmlFor="confirmPassword"
                      className="font-body text-sm font-semibold text-[#1B4332]"
                    >
                      Confirmar senha
                    </label>
                    <div className="relative">
                      <Lock className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-[#2D6A4F]" />
                      <Input
                        id="confirmPassword"
                        name="confirmPassword"
                        type="password"
                        placeholder="Confirme sua senha"
                        value={formData.confirmPassword}
                        onChange={handleInputChange}
                        required={!isLogin}
                        className="pl-12 font-body border-2 border-[#1B4332]/10 rounded-xl h-12 focus:border-[#52B788]"
                      />
                    </div>
                  </div>
                )}

                {isLogin && (
                  <div className="flex items-center justify-between text-sm font-body">
                    <label className="flex items-center space-x-2 cursor-pointer">
                      <input
                        type="checkbox"
                        className="rounded border-2 border-[#1B4332]/20 text-[#52B788] focus:ring-[#52B788]"
                      />
                      <span className="text-[#2D6A4F]">Lembrar-me</span>
                    </label>
                    <Link
                      to="/forgot-password"
                      className="text-[#52B788] hover:text-[#40916C] font-semibold transition-colors"
                    >
                      Esqueceu a senha?
                    </Link>
                  </div>
                )}

                <Button
                  type="submit"
                  className="w-full btn-primary-custom font-body text-lg font-semibold rounded-full h-14 mt-6"
                  disabled={isLoading}
                >
                  {isLoading ? "Carregando..." : isLogin ? "Entrar" : "Criar conta"}
                </Button>
              </form>

              <div className="relative my-6">
                <div className="absolute inset-0 flex items-center">
                  <Separator className="bg-[#1B4332]/20" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                  <span className="bg-white px-4 text-[#2D6A4F] font-body font-semibold tracking-wider">
                    Ou continue com
                  </span>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <Button
                  variant="outline"
                  className="font-body border-2 border-[#1B4332]/20 hover:border-[#52B788] hover:bg-[#52B788]/5 rounded-xl h-12"
                >
                  <svg className="mr-2 h-5 w-5" viewBox="0 0 24 24">
                    <path
                      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                      fill="#4285F4"
                    />
                    <path
                      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                      fill="#34A853"
                    />
                    <path
                      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                      fill="#FBBC05"
                    />
                    <path
                      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                      fill="#EA4335"
                    />
                  </svg>
                  Google
                </Button>
                <Button
                  variant="outline"
                  className="font-body border-2 border-[#1B4332]/20 hover:border-[#52B788] hover:bg-[#52B788]/5 rounded-xl h-12"
                >
                  <svg className="mr-2 h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" />
                  </svg>
                  Facebook
                </Button>
              </div>
            </CardContent>
          </Card>

          <div className="text-center mt-8 font-body">
            <p className="text-[#2D6A4F]">
              {isLogin ? "Não tem uma conta?" : "Já tem uma conta?"}{" "}
              <button
                onClick={() => setIsLogin(!isLogin)}
                className="text-[#52B788] hover:text-[#40916C] font-bold transition-colors"
              >
                {isLogin ? "Cadastre-se" : "Entre"}
              </button>
            </p>
          </div>

          <div className="text-center mt-6">
            <Link
              to="/"
              className="inline-flex items-center font-body text-[#2D6A4F] hover:text-[#52B788] transition-colors font-semibold"
            >
              ← Voltar ao início
            </Link>
          </div>
        </div>
      </div>
    </>
  );
};

export default Login;
