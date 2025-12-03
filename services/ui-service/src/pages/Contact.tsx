import { zodResolver } from "@hookform/resolvers/zod";
import { Clock, Mail, MapPin, Phone } from "lucide-react";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import Header from "@/components/Header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "@/hooks/use-toast";
import { designSystemStyles } from "@/styles/design-system";

const formSchema = z.object({
  name: z.string().min(2, {
    message: "Nome deve ter pelo menos 2 caracteres.",
  }),
  email: z.string().email({
    message: "Por favor, insira um email válido.",
  }),
  message: z.string().min(10, {
    message: "A mensagem deve ter pelo menos 10 caracteres.",
  }),
});

const Contact = () => {
  const [isVisible, setIsVisible] = useState(false);
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      email: "",
      message: "",
    },
  });

  useEffect(() => {
    setIsVisible(true);
  }, []);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add("animate-in");
          }
        });
      },
      { threshold: 0.1 }
    );

    const elements = document.querySelectorAll(".observe-animation");
    elements.forEach((element) => observer.observe(element));

    return () => observer.disconnect();
  }, []);

  function onSubmit(values: z.infer<typeof formSchema>) {
    const subject = encodeURIComponent("Novo contato via Site Velure");
    const body = encodeURIComponent(
      `Nome: ${values.name}\nEmail: ${values.email}\n\nMensagem:\n${values.message}`
    );

    window.location.href = `mailto:israelschroederm@gmail.com?subject=${subject}&body=${body}`;

    toast({
      title: "Mensagem preparada!",
      description: "Seu cliente de email será aberto para enviar a mensagem.",
    });

    form.reset();
  }

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-[#F8FAF5]">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          <div className="max-w-6xl mx-auto">
            <div className={`text-center mb-16 ${isVisible ? 'hero-enter active' : 'hero-enter'}`}>
              <span className="font-body text-[#52B788] font-semibold text-sm tracking-widest uppercase mb-4 block">
                Fale Conosco
              </span>
              <h1 className="font-display text-5xl lg:text-6xl font-bold text-[#1B4332] mb-6">
                Entre em contato
              </h1>
              <div className="w-20 h-1 bg-gradient-to-r from-[#52B788] to-[#A7C957] mx-auto mb-6" />
              <p className="font-body text-xl text-[#2D6A4F] max-w-2xl mx-auto">
                Tem dúvidas sobre nossos produtos ou precisa de ajuda com seu pet? Estamos aqui para
                ajudar!
              </p>
            </div>

            <div className="grid lg:grid-cols-2 gap-12">
              {/* Contact Form */}
              <Card className="shadow-2xl border-2 border-[#1B4332]/10 rounded-3xl observe-animation">
                <CardHeader className="pt-8 px-8">
                  <CardTitle className="font-display text-3xl font-bold text-[#1B4332]">
                    Envie-nos uma mensagem
                  </CardTitle>
                  <CardDescription className="font-body text-base text-[#2D6A4F] mt-2">
                    Preencha o formulário abaixo e retornaremos em breve.
                  </CardDescription>
                </CardHeader>
                <CardContent className="px-8 pb-8">
                  <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                      <FormField
                        control={form.control}
                        name="name"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel className="font-body font-semibold text-[#1B4332]">
                              Nome
                            </FormLabel>
                            <FormControl>
                              <Input
                                placeholder="Seu nome"
                                {...field}
                                className="font-body border-2 border-[#1B4332]/10 rounded-xl h-12 focus:border-[#52B788]"
                              />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="email"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel className="font-body font-semibold text-[#1B4332]">
                              E-mail
                            </FormLabel>
                            <FormControl>
                              <Input
                                placeholder="seu@email.com"
                                {...field}
                                className="font-body border-2 border-[#1B4332]/10 rounded-xl h-12 focus:border-[#52B788]"
                              />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="message"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel className="font-body font-semibold text-[#1B4332]">
                              Mensagem
                            </FormLabel>
                            <FormControl>
                              <Textarea
                                placeholder="Como podemos ajudar?"
                                className="resize-none font-body border-2 border-[#1B4332]/10 rounded-xl focus:border-[#52B788]"
                                rows={5}
                                {...field}
                              />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <Button
                        type="submit"
                        className="w-full btn-primary-custom font-body text-lg font-semibold rounded-full h-14 mt-2"
                      >
                        Enviar mensagem
                      </Button>
                    </form>
                  </Form>
                </CardContent>
              </Card>

              {/* Contact Information */}
              <div className="space-y-6">
                {[
                  {
                    icon: MapPin,
                    title: "Endereço",
                    content: ["Rua dos Pets, 123", "São Paulo, SP 01234-567"],
                    color: "from-[#52B788] to-[#40916C]",
                    delay: "0s",
                  },
                  {
                    icon: Phone,
                    title: "Telefone",
                    content: ["(11) 1234-5678", "(11) 98765-4321"],
                    color: "from-[#95D5B2] to-[#2D6A4F]",
                    delay: "0.1s",
                  },
                  {
                    icon: Mail,
                    title: "Email",
                    content: ["info@velure.pet", "support@velure.pet"],
                    color: "from-[#A7C957] to-[#E5B520]",
                    delay: "0.2s",
                  },
                  {
                    icon: Clock,
                    title: "Horário",
                    content: [
                      "Segunda - Sexta: 9h - 20h",
                      "Sábado: 9h - 18h",
                      "Domingo: 10h - 17h",
                    ],
                    color: "from-[#52B788] to-[#40916C]",
                    delay: "0.3s",
                  },
                ].map((item, index) => {
                  const Icon = item.icon;
                  return (
                    <Card
                      key={index}
                      className="shadow-lg border-2 border-[#1B4332]/10 rounded-3xl card-hover-subtle observe-animation"
                      style={{ animationDelay: item.delay }}
                    >
                      <CardContent className="p-6">
                        <div className="flex items-start space-x-5">
                          <div
                            className={`bg-gradient-to-br ${item.color} rounded-2xl p-4 flex-shrink-0`}
                          >
                            <Icon className="h-7 w-7 text-white" />
                          </div>
                          <div>
                            <h3 className="font-display text-xl font-bold text-[#1B4332] mb-2">
                              {item.title}
                            </h3>
                            <div className="font-body text-[#2D6A4F] space-y-1">
                              {item.content.map((line, i) => (
                                <p key={i}>{line}</p>
                              ))}
                            </div>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  );
                })}
              </div>
            </div>
          </div>
        </main>
      </div>
    </>
  );
};

export default Contact;
