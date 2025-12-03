import { zodResolver } from "@hookform/resolvers/zod";
import { Clock, Mail, MapPin, Phone } from "lucide-react";
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
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      email: "",
      message: "",
    },
  });

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
    <div className="min-h-screen bg-background">
      <Header />

      <main className="container mx-auto px-4 py-12">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h1 className="text-4xl font-bold text-foreground mb-4">Entre em contato</h1>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
              Tem dúvidas sobre nossos produtos ou precisa de ajuda com seu pet? Estamos aqui para
              ajudar!
            </p>
          </div>

          <div className="grid lg:grid-cols-2 gap-12">
            {/* Contact Form */}
            <Card className="shadow-soft">
              <CardHeader>
                <CardTitle className="text-primary">Envie-nos uma mensagem</CardTitle>
                <CardDescription>
                  Preencha o formulário abaixo e retornaremos em breve.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Form {...form}>
                  <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                    <FormField
                      control={form.control}
                      name="name"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Nome</FormLabel>
                          <FormControl>
                            <Input
                              placeholder="Seu nome"
                              {...field}
                              className="border-border focus:ring-primary"
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
                          <FormLabel>E-mail</FormLabel>
                          <FormControl>
                            <Input
                              placeholder="seu@email.com"
                              {...field}
                              className="border-border focus:ring-primary"
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
                          <FormLabel>Mensagem</FormLabel>
                          <FormControl>
                            <Textarea
                              placeholder="Como podemos ajudar?"
                              className="resize-none border-border focus:ring-primary"
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
                      className="w-full bg-gradient-primary hover:opacity-90 text-primary-foreground"
                    >
                      Enviar mensagem
                    </Button>
                  </form>
                </Form>
              </CardContent>
            </Card>

            {/* Contact Information */}
            <div className="space-y-6">
              <Card className="shadow-soft">
                <CardContent className="p-6">
                  <div className="flex items-start space-x-4">
                    <div className="bg-primary rounded-lg p-3">
                      <MapPin className="h-6 w-6 text-primary-foreground" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-foreground">Endereço</h3>
                      <p className="text-muted-foreground mt-1">
                        Rua dos Pets, 123
                        <br />
                        São Paulo, SP 01234-567
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-soft">
                <CardContent className="p-6">
                  <div className="flex items-start space-x-4">
                    <div className="bg-secondary rounded-lg p-3">
                      <Phone className="h-6 w-6 text-secondary-foreground" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-foreground">Telefone</h3>
                      <p className="text-muted-foreground mt-1">
                        (11) 1234-5678
                        <br />
                        (11) 98765-4321
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-soft">
                <CardContent className="p-6">
                  <div className="flex items-start space-x-4">
                    <div className="bg-accent rounded-lg p-3">
                      <Mail className="h-6 w-6 text-accent-foreground" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-foreground">Email</h3>
                      <p className="text-muted-foreground mt-1">
                        info@velure.pet
                        <br />
                        support@velure.pet
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-soft">
                <CardContent className="p-6">
                  <div className="flex items-start space-x-4">
                    <div className="bg-primary rounded-lg p-3">
                      <Clock className="h-6 w-6 text-primary-foreground" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-foreground">Horário</h3>
                      <div className="text-muted-foreground mt-1">
                        <p>Segunda - Sexta: 9h - 20h</p>
                        <p>Sábado: 9h - 18h</p>
                        <p>Domingo: 10h - 17h</p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default Contact;
