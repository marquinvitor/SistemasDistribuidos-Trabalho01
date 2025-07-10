package com.ejbst;

import jakarta.ejb.Stateful;

// EJB Stateless que retorna uma mensagem
@Stateful
public class HelloService {
    public String sayHello(String name) {
        return "Olá, " + name + "! Bem-vindo à API EJB.";
    }
}