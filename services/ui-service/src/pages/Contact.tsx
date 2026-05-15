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
    message: "Name must be at least 2 characters.",
  }),
  email: z.string().email({
    message: "Please enter a valid email.",
  }),
  message: z.string().min(10, {
    message: "Message must be at least 10 characters.",
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
    elements.forEach((element) => {
      observer.observe(element);
    });

    return () => observer.disconnect();
  }, []);

  function onSubmit(values: z.infer<typeof formSchema>) {
    const subject = encodeURIComponent("New contact from Velure site");
    const body = encodeURIComponent(
      `Name: ${values.name}\nEmail: ${values.email}\n\nMessage:\n${values.message}`
    );

    window.location.href = `mailto:israelschroederm@gmail.com?subject=${subject}&body=${body}`;

    toast({
      title: "Message ready!",
      description: "Your email client will open to send the message.",
    });

    form.reset();
  }

  return (
    <>
      <style>{designSystemStyles}</style>
      <div className="min-h-screen bg-white">
        <Header />

        <main className="container mx-auto px-4 lg:px-8 py-12">
          <div className="max-w-6xl mx-auto">
            <div className={`text-center mb-16 ${isVisible ? "hero-enter active" : "hero-enter"}`}>
              <span className="font-body text-[#52B788] font-semibold text-sm tracking-widest uppercase mb-4 block">
                Contact us
              </span>
              <h1 className="font-display text-5xl lg:text-6xl font-bold text-[#1B4332] mb-6">
                Get in touch
              </h1>
              <div className="w-20 h-1 bg-gradient-to-r from-[#52B788] to-[#A7C957] mx-auto mb-6" />
              <p className="font-body text-xl text-[#2D6A4F] max-w-2xl mx-auto">
                Have questions about our products or need help with your pet? We're here to help!
              </p>
            </div>

            <div className="grid lg:grid-cols-2 gap-12">
              {/* Contact Form */}
              <Card className="shadow-2xl border-2 border-[#1B4332]/10 rounded-3xl observe-animation">
                <CardHeader className="pt-8 px-8">
                  <CardTitle className="font-display text-3xl font-bold text-[#1B4332]">
                    Send us a message
                  </CardTitle>
                  <CardDescription className="font-body text-base text-[#2D6A4F] mt-2">
                    Fill out the form below and we'll get back to you soon.
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
                              Name
                            </FormLabel>
                            <FormControl>
                              <Input
                                placeholder="Your name"
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
                              Email
                            </FormLabel>
                            <FormControl>
                              <Input
                                placeholder="you@email.com"
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
                              Message
                            </FormLabel>
                            <FormControl>
                              <Textarea
                                placeholder="How can we help?"
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
                        Send message
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
                    title: "Address",
                    content: ["123 Pet Street", "San Francisco, CA 94105"],
                    color: "from-[#52B788] to-[#40916C]",
                    delay: "0s",
                  },
                  {
                    icon: Phone,
                    title: "Phone",
                    content: ["+1 (415) 555-1234", "+1 (415) 555-5678"],
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
                    title: "Hours",
                    content: [
                      "Monday – Friday: 9am – 8pm",
                      "Saturday: 9am – 6pm",
                      "Sunday: 10am – 5pm",
                    ],
                    color: "from-[#52B788] to-[#40916C]",
                    delay: "0.3s",
                  },
                ].map((item) => {
                  const Icon = item.icon;
                  return (
                    <Card
                      key={item.title}
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
                              {item.content.map((line) => (
                                <p key={line}>{line}</p>
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
